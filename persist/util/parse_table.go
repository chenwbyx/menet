package util

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

const (
	EWarningFlagUnloadFunction = 1 << iota
	EWarningFlagMax
)

const (
	// 保证所有的索引值完全一致, 不开启则保证最终一致性(不建议开启!!! 串行化所有的索引操作, 影响并发量)
	EOptimizeFlagIndexMutex = 1 << iota
	// 保证数据库索引关系完全一致, 不开启则保证最终一致性(建议开启!!! 内存中模拟合并, 性能影响不大, 减轻数据库开销)
	EOptimizeFlagSQLMerge
	// 使用生成代码取代反射, 固定的模型写法, 不保证所有情况正确(建议开启!!! 解决20%的反射开销, 生成代码不一定考虑到了所有情况, 特殊写法需要先测试)
	EOptimizeFlagGenerateCode
	// 合并插入语句, 减少数据库IO, 减少反射次数(建议开启!!! 插入语句越多, 性能提升越明显, 插入非常少,可以不开启,减少cpu消耗)
	EOptimizeFlagSQLInsertMerge
	EOptimizeFlagMax
)

// {{index .HashIndexUnload.Cols 0}} {{index .HashIndexUnload.Types 0}}

type Index struct {
	Cols            []string
	Keys            string
	ClsPointKeys    string
	CommaKeys       string
	CommaKeyKeys    string
	KeyTypes        string
	Types           []string
	Unique          bool
	Pk              bool
	KeyTypesStripPk string
	KeysStripPk     []string
	EffectIndex     []*Index
}

func (i *Index) String() string {
	return fmt.Sprintf("{Index Cols=%s, Types=%s, Unique=%t, Pk=%t}\t", i.Cols, i.Types, i.Unique, i.Pk)
}

func (i *Index) StringDetail() string {
	return fmt.Sprintf(`{Index
 Cols=%s,
 Keys=%s,
 ClsPointKeys=%s,
 CommaKeys=%s,
 KeyTypes=%s,
 Types=%s,
 Unique=%t,
 Pk=%t
}`, i.Cols, i.Keys, i.ClsPointKeys, i.CommaKeys,
		i.KeyTypes, i.Types, i.Unique, i.Pk)
}

type ArgsInfo struct {
	Save           bool
	PersistPkgPath string
	PersistPkgName string
	UnloadKey      string
	Unload         bool
	WarningFlag    int64
	BombDir        string
	OptimizeFlag   int64
	MaxInsertRows  int64
	QueueThreshold int64
	FileName       string
}

func (args *ArgsInfo) OptimizeFlagIndexMutex() bool {
	return args.OptimizeFlag&EOptimizeFlagIndexMutex != 0
}
func (args *ArgsInfo) OptimizeFlagSQLMerge() bool {
	return args.OptimizeFlag&EOptimizeFlagSQLMerge != 0
}
func (args *ArgsInfo) OptimizeFlagGenerateCode() bool {
	return args.OptimizeFlag&EOptimizeFlagGenerateCode != 0
}
func (args *ArgsInfo) OptimizeFlagSQLInsertMerge() bool {
	return args.OptimizeFlag&EOptimizeFlagSQLInsertMerge != 0
}

func ValidWarningFlag(value, flag int64) bool {
	return (value & flag) == 1
}

// indexList [{"cols":[key1, key2], "types":[type1, type2], "unique"=true "pk"=true},... ]
type MyVisitor struct {
	Depth               int
	DataName            string
	PackageName         string
	DataTypeSpec        *ast.TypeSpec
	DataStructType      *ast.StructType
	HashIndexList       []*Index
	ModifyHashIndexList []*Index
	ModifyColList       []string
	ColEffectHashMap    map[string][]*Index
	KeyTypeMap          map[string]string
	HashIndexUnload     *Index
	HashIndexPk         *Index
	ArgsInfo            *ArgsInfo
}

func (v *MyVisitor) HasGlobalFuncNewPersist() bool {
	has := true
	if v.HashIndexPk != nil && v.ArgsInfo.Save && v.HashIndexUnload != nil {
		exist := false
		for _, col := range v.HashIndexPk.Cols {
			if col == v.HashIndexUnload.Cols[0] {
				exist = true
			}
		}
		if !exist {
			has = false
		}
	}
	return has
}

func (v *MyVisitor) ReverseHashIndexList() (lst []*Index) {
	for i := len(v.HashIndexList) - 1; i >= 0; i-- {
		lst = append(lst, v.HashIndexList[i])
	}
	return
}

func ParseTable(filePathAbs string, tableName string, argsInfo *ArgsInfo) (visitor *MyVisitor) {

	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, filePathAbs, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		fmt.Println(err.Error())
	}
	//ast.Print(fs, f)
	visitor = &MyVisitor{}
	visitor.DataName = tableName
	visitor.ArgsInfo = argsInfo

	ast.Walk(visitor, f)

	//fmt.Println(visitor)

	indexDict := make(map[string]*Index)
	var keyNameList []string

	visitor.KeyTypeMap = make(map[string]string)
	visitor.ColEffectHashMap = make(map[string][]*Index)

	// 遍历所有的tag, 提取索引信息
	for _, field := range GetStructBuildInTypeFields(visitor.DataStructType) {
		if _, ok := field.Type.(*ast.Ident); !ok {
			continue
		}
		//fmt.Println(GetFieldName(field))
		//fmt.Println(GetFieldType(field))
		//fmt.Println(GetFieldTags(field))
		var isPk bool
		var isUnique bool
		var index *Index
		var ok bool
		for _, tag := range GetFieldTags(field).Tags() {
			if tag.Key == "xorm" && strings.Contains(tag.Name, "pk") {
				isPk = true
			}
		}

		for _, tag := range GetFieldTags(field).Tags() {
			index = nil
			keyName := tag.Key
			if strings.Contains(tag.Key, "hash") {
				items := strings.Split(tag.Name, ";")
				for _, item := range items {
					kvs := strings.Split(item, "=")
					switch kvs[0] {
					case "unique":
						isUnique = kvs[1] == "1"
					case "pk":
						isPk = kvs[1] == "1"
					case "group":
						keyName += kvs[1]
					}
				}
				if index, ok = indexDict[keyName]; !ok {
					keyNameList = append(keyNameList, keyName)
					index = &Index{}
					indexDict[keyName] = index
				}
				if _, ok = visitor.KeyTypeMap[GetFieldName(field)]; !ok {
					visitor.KeyTypeMap[GetFieldName(field)] = GetFieldType(field)
					visitor.ModifyColList = append(visitor.ModifyColList, GetFieldName(field))
				}

				index.Cols = append(index.Cols, GetFieldName(field))
				index.Keys += GetFieldName(field)
				index.ClsPointKeys += "cls." + GetFieldName(field) + ","
				index.CommaKeys += GetFieldName(field) + ","
				index.CommaKeyKeys += GetFieldName(field) + ": " + GetFieldName(field) + ","
				index.KeyTypes += GetFieldName(field) + " " + GetFieldType(field) + ","
				index.Types = append(index.Types, GetFieldType(field))
				if index != nil {
					index.Pk = isPk
					index.Unique = isUnique
				}
			} else {
			}
			//fmt.Println(tag.Key, "@", tag.Name)
		}
	}
	//fmt.Println(keyNameList)

	// 主键强制放到第一个
	for _, keyName := range keyNameList {
		index := indexDict[keyName]
		// 记录主键信息
		if index.Pk && index.Unique && visitor.HashIndexPk == nil {
			visitor.HashIndexPk = index
			// 主键移除修改列表
			for _, col := range index.Cols {
				for i := range visitor.ModifyColList {
					if visitor.ModifyColList[i] == col {
						visitor.ModifyColList = append(visitor.ModifyColList[:i], visitor.ModifyColList[i+1:]...)
						break
					}
				}
			}
		} else {
			// 处理非主键的索引修改
			visitor.HashIndexList = append(visitor.HashIndexList, index)
		}
	}
	funcNotInIndex := func(col string, index *Index) bool {
		exist := false
		for _, colPk := range index.Cols {
			if col == colPk {
				exist = true
			}
		}
		return !exist
	}

	for _, index := range visitor.HashIndexList {
		notInPk := false
		for idx, col := range index.Cols {
			if funcNotInIndex(col, visitor.HashIndexPk) {
				index.KeyTypesStripPk += col + " " + index.Types[idx] + ","
				index.KeysStripPk = append(index.KeysStripPk, col)
				notInPk = true
			} else {
			}
		}
		if notInPk {
			if len(index.Cols) == 1 {
				for i := range visitor.ModifyColList {
					if visitor.ModifyColList[i] == index.Cols[0] {
						visitor.ModifyColList = append(visitor.ModifyColList[:i], visitor.ModifyColList[i+1:]...)
						break
					}
				}
			}
			visitor.ModifyHashIndexList = append(visitor.ModifyHashIndexList, index)
		}
	}
	if visitor.HashIndexPk != nil {
		visitor.HashIndexList = append([]*Index{visitor.HashIndexPk}, visitor.HashIndexList...)
	}

	for _, index := range visitor.HashIndexList {
		for _, otherIndex := range visitor.HashIndexList {
			for _, col := range index.KeysStripPk {
				if !funcNotInIndex(col, otherIndex) {
					index.EffectIndex = append(index.EffectIndex, otherIndex)
					break
				}
			}
		}
	}
	for _, col := range visitor.ModifyColList {
		for _, otherIndex := range visitor.HashIndexList {
			if !funcNotInIndex(col, otherIndex) {
				visitor.ColEffectHashMap[col] = append(visitor.ColEffectHashMap[col], otherIndex)
			}
		}
	}

	//// 主键强制放到第一个
	//for _, v := range indexDict {
	//	if v.Pk && v.Unique {
	//		visitor.HashIndexList = append(visitor.HashIndexList, v)
	//	}
	//}
	//for _, v := range indexDict {
	//	if !(v.Pk && v.Unique) {
	//		visitor.HashIndexList = append(visitor.HashIndexList, v)
	//	}
	//}

	// 查找是否存在Unload key索引
	if argsInfo.Unload {
		for _, keyName := range keyNameList {
			index := indexDict[keyName]
			// 必须存在数据库索引, 不然导入数据很慢
			if len(index.Cols) == 1 && index.Cols[0] == argsInfo.UnloadKey {
				if visitor.HashIndexUnload != nil {
					panic("repeated unload key. " + tableName + index.String())
				}
				visitor.HashIndexUnload = index
			} else {
				// 暂不支持联合键, 用作导出
			}
		}
	}
	if argsInfo.Unload && ValidWarningFlag(argsInfo.WarningFlag, EWarningFlagUnloadFunction) && visitor.HashIndexUnload == nil {
		fmt.Println("warring: can't not generate method load. " + tableName + argsInfo.UnloadKey)
		//for _, v := range unloadWarringIndexList {
		//	fmt.Println("warring: can't not generate method load. " + tableName + v.String())
		//}
	}
	//fmt.Println(visitor.HashIndexUnload)
	//	Uid
	//	int32
	//xorm:"pk" hash:"group=1;unique=0" hash:"group=2;unique=1"
	//	Id
	//	int32
	//xorm:"pk" hash:"group=2;unique=1"

	//for _, field := range GetStructBuildOutTypeFields(visitor.DataStructType) {
	//	block := GetFiledDeepCopyString(field)
	//	fmt.Println("##########",block)
	//}

	return
}

func (v *MyVisitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		v.Depth -= 1
		return nil
	}
	ts, ok := n.(*ast.TypeSpec)
	if ok && ts.Name.Name == v.DataName {
		v.DataTypeSpec = ts
		if structType, ok := ts.Type.(*ast.StructType); ok {
			v.DataStructType = structType
		}
	}

	//var s string
	//switch node := n.(type) {
	//case *ast.Ident:
	//	s = node.Name
	//case *ast.BasicLit:
	//	s = node.Value
	//}
	//fmt.Printf("%s%T: %s\n", strings.Repeat("\t", int(v.Depth)), n, s)
	v.Depth += 1
	return v
}
