package main

import (
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type commitInfo struct {
	User    string `json:"user"`
	Task    string `json:"task"`
	Story   string `json:"story"`
	Bug     string `json:"bug"`
	Message string `json:"message"`
}

func main() {
	currentBranch, _ := exec.Command("git", "symbolic-ref", "--short", "HEAD").CombinedOutput()
	currentBranchName := strings.Trim(string(currentBranch), " ")
	if strings.Contains(currentBranchName, "fatal: not a git repository") {
		errors.New("当前目录没有git")
	}

	fmt.Println("当前分支:", currentBranchName)

	doCommit(currentBranchName)

}

func getCommitInfo() (string, error) {
	defaultUserName, _ := exec.Command("git", "config", "user.name").CombinedOutput()

	var c commitInfo

	flag.StringVar(&c.User, "u", strings.Trim(string(defaultUserName), "\n"), "--user=[user nick](默认为git config user.name)")
	flag.StringVar(&c.Task, "t", "", "--task=[task id]")
	flag.StringVar(&c.Story, "s", "", "--story=[story id]")
	flag.StringVar(&c.Bug, "b", "", "--bug=[bug id]")
	flag.StringVar(&c.Message, "m", "", "提交信息")
	flag.Parse()

	commitInfo := "--user=" + c.User

	var isTapd, isCommitInfo = false, false

	if c.Task != "" {
		isTapd = true
		commitInfo += " --task=" + c.Task
	}
	if c.Story != "" {
		isTapd = true
		commitInfo += " --story=" + c.Story
	}
	if c.Bug != "" {
		isTapd = true
		commitInfo += " --bug=" + c.Bug
	}

	if c.Message != "" {
		isCommitInfo = true
		commitInfo += " " + c.Message
	}

	if !isTapd {
		errors.New("请输入关联TAPD信息")
	}

	if !isCommitInfo {
		errors.New("请输入提交信息")
	}
	fmt.Println("确认提交信息：" + commitInfo)
	res := scanLn("1：确认 其他:取消")
	if res == "1" {
		return commitInfo, nil
	}
	return "", errors.New("你已经取消")

}

func doCommit(currentBranchName string) {
	commitType := scanLn("请输入提交模式 1)开发分支+版本分支+目标分支  2)提交当前分支  3)分支合并到另一分支 : ")
	ct, _ := strconv.Atoi(commitType)
	switch ct {
	case 1:
		versionBranchName := scanLn("请输入版本分支:")
		checkBranchExist(versionBranchName)
		targetBranchName := scanLn("请输入目标分支:")
		checkBranchExist(targetBranchName)
		info, err := getCommitInfo()
		add(info)
		pushBranch(currentBranchName)

		checkOutBranch(versionBranchName)
		pullBranch(versionBranchName)
		mergeBranch(currentBranchName)
		pushBranch(versionBranchName)

		checkOutBranch(targetBranchName)
		pullBranch(targetBranchName)
		mergeBranch(versionBranchName)
		pushBranch(targetBranchName)

	case 2:
		info := getCommitInfo()
		add(info)
		pushBranch(currentBranchName)
	case 3:
		needMerge := scanLn("请输入需要合并的分支:")
		checkBranchExist(needMerge)
		targetBranchName := scanLn("请输入目标分支:")
		checkBranchExist(targetBranchName)

		pullBranch(needMerge)

		checkOutBranch(targetBranchName)
		pullBranch(targetBranchName)
		mergeBranch(needMerge)
		pushBranch(targetBranchName)

	default:
		fmt.Println("输入不合法")
		doCommit(currentBranchName)
	}
}

func add(info string) {
	fmt.Println("git add .")
	exec.Command("git", "add", ".").Run()
	fmt.Println("git commit -m '" + info + "'")
	exec.Command("git", "commit", "-m", info).Run()
}

func pullBranch(branchName string) {
	fmt.Println("git pull origin " + branchName)
	res, _ := exec.Command("git", "pull", "origin", branchName).CombinedOutput()
	strRes := strings.Trim(string(res), "\n")
	fmt.Println(strRes)
	if strings.Contains(strRes, "Automatic merged failed") {
		errors.New("当前分支有冲突")
	}
	if strings.Contains(strRes, "fatal: invalid") {
		errors.New("没有权限")
	}
}

func pushBranch(branchName string) {
	fmt.Println("git push origin " + branchName)
	res, _ := exec.Command("git", "push", "origin", branchName).CombinedOutput()
	strRes := strings.Trim(string(res), "\n")
	if strings.Contains(strRes, "gti pull ...") {
		pullBranch(branchName)
		pushBranch(branchName)
	}
}

func mergeBranch(branchName string) {
	fmt.Println("git merge " + branchName)
	res, _ := exec.Command("git", "merge", branchName).CombinedOutput()
	strRes := strings.Trim(string(res), "\n")
	if strings.Contains(strRes, "Automatic merged failed") {
		errors.New("合并到分支有冲突需要手动合并")
	}
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
		return commitType
	}
}

//判断分支是否存在
func checkBranchExist(branchName string) {
	res, _ := exec.Command("git", "rev-parse", "--verify", branchName).CombinedOutput()
	strRes := strings.Trim(string(res), "\n")
	if strRes == "fatal: Needed a single revision" {
		errors.New("分支：" + branchName + "不存在")
	}
}
