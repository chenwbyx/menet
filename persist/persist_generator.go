//go:generate go install menet/persist
package main

import (
	"bytes"
	"flag"
	"fmt"
	_ "github.com/getlantern/deepcopy"
	_ "github.com/wbexwbex/queue"
	"io/ioutil"
	"log"
	"math"
	tpl "menet/persist/template"
	"menet/persist/util"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"text/template"
)

func GenPersistMethod(wg *sync.WaitGroup, srcDir, goFile, dstDir, goPackage, tableName string, argsInfo *util.ArgsInfo) {
	defer wg.Done()
	tt := template.Must(template.New("Main").Parse(tpl.Main))
	//fmt.Print(tableName, " ,")
	genFileName := "001_" + strings.ToLower(tableName) + "_persist.go"

	goFilePathAbs := path.Join(srcDir, goFile)
	visitor := util.ParseTable(goFilePathAbs, tableName, argsInfo)
	visitor.PackageName = goPackage

	//fmt.Println(visitor)

	// 处理格式并调整所提供文件的导入
	var fileBuf = &bytes.Buffer{}
	filePath := path.Join(dstDir, genFileName)
	err := tt.Execute(fileBuf, visitor)
	if err != nil {
		panic(err.Error())
	}
	out := fileBuf.Bytes()
	//不能并发执行
	//out, err := imports.Process(filePath, fileBuf.Bytes(), nil)
	//if err != nil {
	//	panic(err.Error())
	//}
	err = ioutil.WriteFile(filePath, out, 0666)
	if err != nil {
		panic(err.Error())
	}

	err = exec.Command("goimports", "-w", filePath).Run()
	if err != nil {
		panic(err.Error())
	}

	genPersistSet := false
	// 生成索引依赖的sync.Map
	if visitor.HashIndexUnload != nil {
		name := visitor.DataName + "MapUnload"
		filePath := path.Join(dstDir, "001_"+strings.ToLower(visitor.DataName)+"_map_unload.go")
		cmd := exec.Command("syncmap", "-name="+name, "-pkg="+visitor.PackageName, "-o="+filePath, "map["+visitor.HashIndexUnload.Types[0]+"]*int32")
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Panic("exec.syncmap gen persist set failed with", err.Error(), string(out))
		}
	}
	for _, v := range visitor.HashIndexList {
		if !v.Unique {
			if !genPersistSet {
				name := visitor.DataName + "Set"
				filePath := path.Join(dstDir, "001_"+strings.ToLower(visitor.DataName)+"_set.go")
				cmd := exec.Command("syncmap", "-name="+name, "-pkg="+visitor.PackageName, "-o="+filePath, "map[*"+visitor.DataName+"]bool")
				out, err = cmd.CombinedOutput()
				if err != nil {
					log.Panic("exec.syncmap gen persist set failed with", err.Error(), string(out))
				}
				genPersistSet = true
			}
		}
		if len(v.Cols) > 1 {
			if v.Unique {
				name := visitor.DataName + "Hash" + v.Keys
				filePath := path.Join(dstDir, "001_"+strings.ToLower(name)+".go")
				cmd := exec.Command("syncmap", "-name="+name, "-pkg="+visitor.PackageName, "-o="+filePath, "map["+visitor.DataName+"KeyTypeHash"+v.Keys+"]*"+visitor.DataName)
				out, err = cmd.CombinedOutput()
				if err != nil {
					log.Panic("exec.syncmap gen"+name+" failed with", err.Error(), string(out))
				}
			} else {
				name := visitor.DataName + "Hash" + v.Keys
				filePath := path.Join(dstDir, "001_"+strings.ToLower(name)+".go")
				cmd := exec.Command("syncmap", "-name="+name, "-pkg="+visitor.PackageName, "-o="+filePath, "map["+visitor.DataName+"KeyTypeHash"+v.Keys+"]*"+visitor.DataName+"Set")
				out, err = cmd.CombinedOutput()
				if err != nil {
					log.Panic("exec.syncmap gen"+name+" failed with", err.Error(), string(out))
				}
			}
		} else {
			if v.Unique {
				name := visitor.DataName + "Hash" + v.Keys
				filePath := path.Join(dstDir, "001_"+strings.ToLower(name)+".go")
				cmd := exec.Command("syncmap", "-name="+name, "-pkg="+visitor.PackageName, "-o="+filePath, "map["+v.Types[0]+"]*"+visitor.DataName)
				out, err = cmd.CombinedOutput()
				if err != nil {
					log.Panic("exec.syncmap gen"+name+" failed with", err.Error(), string(out))
				}
			} else {
				name := visitor.DataName + "Hash" + v.Keys
				filePath := path.Join(dstDir, "001_"+strings.ToLower(name)+".go")
				cmd := exec.Command("syncmap", "-name="+name, "-pkg="+visitor.PackageName, "-o="+filePath, "map["+v.Types[0]+"]*"+visitor.DataName+"Set")
				out, err = cmd.CombinedOutput()
				if err != nil {
					log.Panic("exec.syncmap gen"+name+" failed with", err.Error(), string(out))
				}
			}
		}
	}
}
func mainT() {
	var dir, goFile, goPackage, tableName string
	dirSrc := "G:/MyProjectGo/ultraman/src/menet/persist/tests/protocol"
	dirDst := "G:/MyProjectGo/ultraman/src/menet/persist/tests/model"
	goFile = "col_type.go"
	goPackage = "model"
	tableName = "Persist"
	bombDir := "./" + strings.Replace(strings.Replace(strings.Replace(dir, "\\", "_", -1), ":", "", -1), "/", "_", -1)
	*unload = true
	persistPkgPath := "menet/persist/tests/protocol"
	persistPkgName := persistPkgPath[strings.LastIndex(persistPkgPath, "/")+1:]
	argsInfo := util.ArgsInfo{
		Save:           *save,
		PersistPkgPath: persistPkgPath,
		PersistPkgName: persistPkgName,
		UnloadKey:      *unloadKey,
		Unload:         *unload,
		WarningFlag:    *warningFlag,
		BombDir:        bombDir,
		OptimizeFlag:   *optimizeFlag,
		MaxInsertRows:  *maxInsertRows,
		QueueThreshold: *queueThreshold,
		FileName:       goFile,
	}
	GenPersistMethod(&sync.WaitGroup{}, dirSrc, goFile, dirDst, goPackage, tableName, &argsInfo)
}

var save = flag.Bool("save", true, "save data to mysql")
var srcDir = flag.String("src", "", "persist struct path")
var dstDir = flag.String("dst", "", "generate file path")
var fileName = flag.String("fileName", "", "persist file name")
var pkgName = flag.String("pkgName", "", "generate package name")
var unloadKey = flag.String("unloadKey", "Uid", "unload key name")
var unload = flag.Bool("unload", false, "can be unloaded")
var warningFlag = flag.Int64("warningFlag", math.MaxInt32, "warning flag")
var maxInsertRows = flag.Int64("maxInsertRows", 100, "max rows")
var queueThreshold = flag.Int64("queueThreshold", 10000, "sync queue threshold")
var defaultOptimizeFlag int64 = (0 & util.EOptimizeFlagIndexMutex) |
	util.EOptimizeFlagSQLMerge |
	util.EOptimizeFlagGenerateCode |
	util.EOptimizeFlagSQLInsertMerge
var optimizeFlag = flag.Int64("optimizeFlag", defaultOptimizeFlag, "optimize flag")

func main() {
	var err error
	var dir, goFile, goPackage, tableName string
	var persistPkgPath, persistPkgName string
	//GOFILE=item.go
	//GOPACKAGE=data
	dir, err = os.Getwd()
	if err != nil {
		fmt.Printf("Could not GetWd(): %s (skip)\n", err)
		return
	}

	flag.Parse()

	//fmt.Println("os.Args[0:]", os.Args[0:])
	//fmt.Println("os.Args[1:]", os.Args[1:])
	//fmt.Println("save", *save)
	//fmt.Println("args", flag.Args())

	//fmt.Println("work dir ", dir)
	goFile = os.Getenv("GOFILE")
	goPackage = os.Getenv("GOPACKAGE")

	goPath := os.Getenv("GOPATH")
	//fmt.Print(goFile + " ")
	//fmt.Println(goPackage)

	// 初始化路径
	if *srcDir == "" {
		*srcDir = dir
	} else {
		persistPkgPath = *srcDir
		persistPkgName = persistPkgPath[strings.LastIndex(persistPkgPath, "/")+1:]
		*srcDir = path.Join(goPath, "src", *srcDir)
	}
	if *dstDir == "" {
		*dstDir = dir
	} else {
		*dstDir = path.Join(goPath, "src", *dstDir)
	}
	if *fileName == "" {
		*fileName = goFile
	}
	if *pkgName == "" {
		*pkgName = goPackage
	}

	bombDir := "./" + strings.Replace(strings.Replace(strings.Replace(*dstDir, "\\", "_", -1), ":", "", -1), "/", "_", -1)

	//fmt.Println(goPath + " goPath ")
	//fmt.Println(*srcDir + " srcDir ")
	//fmt.Println(*dstDir + " dstDir ")
	//fmt.Println(persistPkgPath + " persistPkgPath ")
	//fmt.Println(persistPkgName + " persistPkgName ")

	argsInfo := util.ArgsInfo{
		Save:           *save,
		PersistPkgPath: persistPkgPath,
		PersistPkgName: persistPkgName,
		UnloadKey:      *unloadKey,
		Unload:         *unload,
		WarningFlag:    *warningFlag,
		BombDir:        bombDir,
		OptimizeFlag:   *optimizeFlag,
		MaxInsertRows:  *maxInsertRows,
		QueueThreshold: *queueThreshold,
		FileName:       *fileName,
	}

	wg := &sync.WaitGroup{}
	for _, tableName = range flag.Args() {
		wg.Add(1)
		// 生成persist类
		go GenPersistMethod(wg, *srcDir, *fileName, *dstDir, *pkgName, tableName, &argsInfo)
	}
	wg.Wait()
	//fmt.Println()
}
