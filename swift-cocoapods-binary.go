package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("Swift CocoaPods Binary")
	fmt.Println("")

	podPtr := flag.String("pod", "", "Name of the pod that you would like to have as a binary")
	versionPtr := flag.String("version", "", "The pod version")
	iosVersionPtr := flag.String("ios", "", "The iOS version; Default: 11.0")
	sourcePtr := flag.String("source", "", "The podspec sources (separate with comma if more than one is needed)")
	flag.Parse()
	if *podPtr == "" {
		fmt.Println("Error: Pod name is missing")
		os.Exit(1)
	}
	if *versionPtr == "" {
		fmt.Println("Error: Pod version is missing")
		os.Exit(1)
	}
	iosVersion := "11.0"
	if *iosVersionPtr != "" {
		iosVersion = *iosVersionPtr
	}
	source := ""
	if *sourcePtr != "" {
		source = ""
		sources := strings.SplitN(*sourcePtr, ",", -1)
		for _, s := range sources {
			source += "source '" + s + "'\n"
		}
	}

	checkPreconditions()

	podfile := fmt.Sprintf(`
%s

use_frameworks!
platform :ios, '%s'

pre_install do |installer|
	installer.analysis_result.specifications.each do |s|
		if s.swift_versions.empty?
			s.swift_versions << '4.2'
		end
	end
end

plugin 'cocoapods-rome', dsym: true, configuration: 'Release'

target 'FrameworksToBuild' do
	pod '%s', '%s'
end
	`, source, iosVersion, *podPtr, *versionPtr)

	tmpPath := "./swift-binary-tmp"
	os.Mkdir(tmpPath, 0777)
	err := ioutil.WriteFile(tmpPath+"/Podfile", []byte(podfile), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Creating binary frameworks")
	fmt.Println("Please wait...")

	os.Chdir(tmpPath)
	cmd := exec.Command("pod", "install")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		os.Exit(1)
	}

	frameworks := listFrameworks(tmpPath)
	os.Rename("Rome", *podPtr)
	createArchive(*podPtr, "../"+*podPtr+".zip", frameworks)
	os.Chdir("..")
	os.RemoveAll(tmpPath)

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println("Successfully created archive with all binary frameworks at:")
	fmt.Println(pwd + "/" + *podPtr + ".zip")
	fmt.Println()
}

func checkPreconditions() {
	_, err := exec.LookPath("gem")
	if err != nil {
		fmt.Println("Please install RubyGems to continue")
		os.Exit(1)
	}
	installCocoapodsIfNeeded()
	installCocoapodsRomeIfNeeded()
}

func installCocoapodsIfNeeded() {
	cmd := exec.Command("gem", "list", "^cocoapods$", "-i")
	if err := cmd.Run(); err != nil {
		fmt.Println("Try to install missing gem 'cocoapods'")
		cmd = exec.Command("gem", "i", "cocoapods")
		if err := cmd.Run(); err != nil {
			fmt.Println("Could not install cocoapods")
			os.Exit(1)
		}
		fmt.Println("Installation successful")
		fmt.Println()
	}
}

func installCocoapodsRomeIfNeeded() {
	cmd := exec.Command("gem", "list", "^cocoapods-rome$", "-i")
	if err := cmd.Run(); err != nil {
		fmt.Println("Try to install missing gem 'cocoapods-rome'")
		cmd = exec.Command("gem", "i", "cocoapods-rome")
		if err := cmd.Run(); err != nil {
			fmt.Println("Could not install cocoapods")
			os.Exit(1)
		}
		fmt.Println("Installation successful")
		fmt.Println()
	}
}

func listFrameworks(path string) []string {
	var files []string
	root := path + "/Rome"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".framework") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

func createArchive(source, target string, targetFiles []string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !find(targetFiles, path) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
