package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// composerStruct 是php项目composer.json的解析映射
// 主要的是php部分确定php版本
type composerStruct struct {
	Require struct {
		Php string `json:"php,omitempty,-"`
	} `json:"require,omitempty,-"`
}

func main() {

	filepath, content := loadConfigurations()

	if content == "" {
		log.Fatalf("!!! EMPTY CONFIG [%s] !!!", filepath)
	}

	configs := parseIni(content)

	if _, ok := configs["default"]; !ok {
		log.Fatalln("!!! NOT FOUND [default] CONFIG ITEMS !!!")
	}

	// 默认要使用的php版本
	def := configs["default"]

	if _, ok := configs[def]; !ok {
		log.Fatalln("!!! [default] PHP version not found in config items. !!!")
	}

	// 获取目标php版本
	targetVersion, err := getTargetVersion()
	if err != nil {
		log.Fatalln(err)
	}

	// 统计版本按顺序排序
	versions := make([]string, 0)
	for v, _ := range configs {
		if v == "default" {
			continue
		}
		versions = append(versions, v)
	}
	sort.Sort(versionSort(versions))

	// 默认要使用的php版本
	bin := configs[configs["default"]]
	ver := ""
	// 删除default的配置
	delete(configs, "default")

	// 执行版本判断逻辑
	operatorsRelation := map[string]int{">=": 0, "<=": -1, ">": 1, "<": -1, "=": 0, "^": 0}
	operators := []string{">=", "<=", ">", "<", "=", "^"}
	for _, operator := range operators {
		tmpVer := targetVersion
		check := operatorsRelation[operator]
		if strings.HasPrefix(targetVersion, operator) {
			tmpVer = strings.TrimPrefix(targetVersion, operator)
		}

		for _, nowVer := range versions {
			if compareVersion(nowVer, tmpVer) >= check {
				fmt.Println("detected", nowVer, operator, tmpVer)
				ver = nowVer
				break
			}
		}
		if ver != "" {
			break
		}
	}
	fmt.Println(ver)
	bin = configs[ver]

	if bin == "" {
		log.Fatalln("!!! invalid version detect !!!")
	}

	cmd := exec.Command(bin, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalln(err)
	}

}

// getTargetVersion 获取当前需要执行的PHP版本
// 优先级 ENV > phpver > composer.json
func getTargetVersion() (string, error) {

	// 从环境变量读取php版本
	// 优先级最高
	// PV=7.2 php xxx
	PV := os.Getenv("PV")

	if PV != "" {
		return PV, nil
	}

	if cwd, err := os.Getwd(); err == nil {
		// 检查当前路径有无phpver文件
		// 如果有则从当前路径读取phpver文件并获取版本
		PHPVersionFile := path.Join(cwd, "phpver")
		if fileExists(PHPVersionFile) {
			PHPVersionFileContent := readFile(PHPVersionFile)
			if PHPVersionFileContent == "default" {
				return "", fmt.Errorf("!!! invalid [php version file] sets ")
			}
			return PHPVersionFileContent, nil
		}

		// 检查当前路径有无composer.json文件
		// 如果有则从当前路径读取composer.json文件并获取版本
		ComposerJSONFILE := "composer.json"
		if fileExists(ComposerJSONFILE) {
			ComposerJSONFileContent := readFile(ComposerJSONFILE)
			var composerJsonStruct composerStruct
			err := json.Unmarshal([]byte(ComposerJSONFileContent), &composerJsonStruct)
			if err != nil {
				return "", err
			}
			if composerJsonStruct.Require.Php != "" {
				composerPHPVersion := composerJsonStruct.Require.Php
				return strings.Split(composerPHPVersion, "|")[0], nil
			}
		}
	}

	return "", fmt.Errorf("parse target version fail")
}

func loadConfigurations() (string, string) {
	cfgDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
		return "", ""
	}

	cfgDir = cfgDir + "/.config"
	if !directoryExists(cfgDir) {
		log.Printf("%s not found\n", cfgDir+"/.config")
		return "", ""
	}

	cfgFile := cfgDir + "/php-detect"
	if !fileExists(cfgFile) {
		if err = writeFile(cfgFile, ""); err != nil {
			log.Fatalln(err)
			return cfgFile, ""
		}
	}

	return cfgFile, readFile(cfgFile)
}

func readFile(filepath string) string {

	file, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("readfile err: %s\n", err)
		return ""
	}

	return string(file)

}

func writeFile(filepath string, content string) error {
	return os.WriteFile(filepath, []byte(content), 0644)
}

func directoryExists(filepath string) bool {
	stat, err := os.Stat(filepath)

	if os.IsNotExist(err) {
		return false
	}

	if !stat.IsDir() {
		return false
	}

	return true
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func parseIni(content string) map[string]string {
	ini := make(map[string]string)

	keyValue := regexp.MustCompile(`^(\w.*)\s*=\s*(.*?)\s*$`)

	contents := strings.Split(strings.ReplaceAll(content, "\r", ""), "\n")

	for _, line := range contents {

		if line == "" {
			continue
		}

		result := keyValue.FindStringSubmatch(line)
		if len(result) > 0 {
			key, value := result[1], result[2]
			// fmt.Printf("kv: '%q'\n", result)
			ini[key] = value
			continue
		}
	}

	return ini
}

func compareVersion(version1 string, version2 string) int {
	var res int
	ver1Strs := strings.Split(version1, ".")
	ver2Strs := strings.Split(version2, ".")
	ver1Len := len(ver1Strs)
	ver2Len := len(ver2Strs)
	verLen := ver1Len
	if len(ver1Strs) < len(ver2Strs) {
		verLen = ver2Len
	}
	for i := 0; i < verLen; i++ {
		var ver1Int, ver2Int int
		if i < ver1Len {
			ver1Int, _ = strconv.Atoi(ver1Strs[i])
		}
		if i < ver2Len {
			ver2Int, _ = strconv.Atoi(ver2Strs[i])
		}
		if ver1Int < ver2Int {
			res = -1
			break
		}
		if ver1Int > ver2Int {
			res = 1
			break
		}
	}
	return res
}

type versionSort []string

func (s versionSort) Len() int {
	return len(s)
}
func (s versionSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s versionSort) Less(i, j int) bool {
	return compareVersion(s[i], s[j]) < 0
}
