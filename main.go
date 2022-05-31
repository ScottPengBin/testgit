package main

import (
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var (
	h, e          bool
	u, t, s, b, m string
)

func main() {
	currentBranch, _ := exec.Command("git", "symbolic-ref", "--short", "HEAD").CombinedOutput()
	currentBranchName := strings.Trim(string(currentBranch), "\n")
	if strings.Contains(currentBranchName, "fatal: not a git repository") {
		fmt.Println("当前目录没有git")
		return
	}

	fmt.Println("当前分支:", currentBranchName)

	flag.Parse()

	//帮助信息
	if h {
		flag.Usage()
		return
	}

	err := doCommit(currentBranchName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}

func init() {
	defaultUserName, _ := exec.Command("git", "config", "user.name").CombinedOutput()

	flag.BoolVar(&h, "h", false, "this help")
	flag.StringVar(&u, "u", strings.Trim(string(defaultUserName), "\n"), "--user=[user nick](默认为git config user.name)")
	flag.StringVar(&t, "t", "", "--task=[task id]")
	flag.StringVar(&s, "s", "", "--story=[story id]")
	flag.StringVar(&b, "b", "", "--bug=[bug id]")
	flag.StringVar(&m, "m", "", "提交信息")
	flag.BoolVar(&e, "e", false, "是否排除关联tapd 默认关联")
}

//获取commit 信息
func getCommitInfo() (string, error) {

	commitInfo := "--user=" + u

	var isTapd, isCommitInfo = false, false

	if t != "" {
		isTapd = true
		commitInfo += " --task=" + t
	}
	if s != "" {
		isTapd = true
		commitInfo += " --story=" + s
	}
	if b != "" {
		isTapd = true
		commitInfo += " --bug=" + b
	}

	if m != "" {
		isCommitInfo = true
		if isTapd == true {
			commitInfo += " " + m
		} else {
			commitInfo = m
		}

	}

	if !isTapd && e == false {
		return "", errors.New("请输入关联TAPD信息")
	}

	if !isCommitInfo {
		return "", errors.New("请输入提交信息")
	}
	fmt.Println("确认提交信息：" + commitInfo)
	res := scanLn("1:确认 其他:取消  : ")
	if res == "1" {
		return commitInfo, nil
	}
	return "", errors.New("你已经取消")

}

//根据不同类型提交
func doCommit(currentBranchName string) error {
	commitType := scanLn("请输入提交模式 1)开发分支+版本分支+目标分支  2)提交当前分支  3)分支合并到另一分支 : ")
	ct, _ := strconv.Atoi(commitType)

	switch ct {
	case 1:
		versionBranchName := scanLn("请输入版本分支:")
		err := checkBranchExist(versionBranchName)
		if err != nil {
			return err
		}

		targetBranchName := scanLn("请输入目标分支:")
		err = checkBranchExist(targetBranchName)
		if err != nil {
			return err
		}

		info, coErr := getCommitInfo()
		if coErr != nil {
			return coErr
		}
		add(info)

		err = pushBranch(currentBranchName)
		if err != nil {
			return err
		}

		checkOutBranch(versionBranchName)
		err = pullBranch(versionBranchName)
		if err != nil {
			return err
		}

		err = mergeBranch(currentBranchName)
		if err != nil {
			return err
		}

		err = pushBranch(versionBranchName)
		if err != nil {
			return err
		}

		checkOutBranch(targetBranchName)
		err = pullBranch(targetBranchName)
		if err != nil {
			return err
		}

		err = mergeBranch(versionBranchName)
		if err != nil {
			return err
		}

		err = pushBranch(targetBranchName)
		if err != nil {
			return err
		}

	case 2:
		info, coErr := getCommitInfo()
		if coErr != nil {
			return coErr
		}
		add(info)
		err := pushBranch(currentBranchName)
		if err != nil {
			return err
		}
	case 3:
		needMerge := scanLn("请输入需要合并的分支:")
		err := checkBranchExist(needMerge)
		if err != nil {
			return err
		}
		targetBranchName := scanLn("请输入目标分支:")
		err = checkBranchExist(targetBranchName)
		if err != nil {
			return err
		}

		err = pullBranch(needMerge)
		if err != nil {
			return err
		}

		checkOutBranch(targetBranchName)
		err = pullBranch(targetBranchName)
		if err != nil {
			return err
		}

		err = mergeBranch(needMerge)
		if err != nil {
			return err
		}

		err = pushBranch(targetBranchName)
		if err != nil {
			return err
		}

	default:
		fmt.Println("输入不合法")
		return doCommit(currentBranchName)
	}
	return nil
}

func add(info string) {
	fmt.Println("git add .")
	_ = exec.Command("git", "add", ".").Run()
	fmt.Println("git commit -m '" + info + "'")
	exec.Command("git", "commit", "-m", info).Run()
}

func pullBranch(branchName string) error {
	fmt.Println("git pull origin " + branchName)
	res, _ := exec.Command("git", "pull", "origin", branchName).CombinedOutput()
	strRes := strings.Trim(string(res), "\n")
	fmt.Println(strRes)
	if strings.Contains(strRes, "Automatic merged failed") {
		return errors.New("当前分支有冲突")
	}
	if strings.Contains(strRes, "fatal: invalid") {
		return errors.New("没有权限")
	}
	return nil
}

func pushBranch(branchName string) error {
	fmt.Println("git push origin " + branchName)
	res, _ := exec.Command("git", "push", "origin", branchName).CombinedOutput()
	strRes := strings.Trim(string(res), "\n")
	fmt.Println(strRes)
	if strings.Contains(strRes, "gti pull ...") {
		err := pullBranch(branchName)
		if err != nil {
			return err
		}
		err2 := pushBranch(branchName)
		if err2 != nil {
			return err2
		}
	}
	return nil
}

func mergeBranch(branchName string) error {
	fmt.Println("git merge " + branchName)
	res, _ := exec.Command("git", "merge", branchName).CombinedOutput()
	strRes := strings.Trim(string(res), "\n")
	if strings.Contains(strRes, "Automatic merged failed") {
		return errors.New("合并到分支有冲突需要手动合并")

	}
	if strings.Contains(strRes, "error") {
		return errors.New("合并到分支有冲突需要手动合并")

	}
	return nil
}

func checkOutBranch(branchName string) {
	fmt.Println("git checkout " + branchName)
	exec.Command("git", "checkout", branchName).Run()
}

func scanLn(message string) string {
	var commitType string
	fmt.Print(message)
	fmt.Scanln(&commitType)
	if commitType == "" {
		return scanLn(message)
	} else {
		return strings.Trim(commitType, "\n")
	}
}

//判断分支是否存在
func checkBranchExist(branchName string) error {
	res, _ := exec.Command("git", "rev-parse", "--verify", branchName).CombinedOutput()
	strRes := strings.Trim(string(res), "\n")
	if strRes == "fatal: Needed a single revision" {
		return errors.New("分支：" + branchName + "不存在")
	}
	return nil
}
