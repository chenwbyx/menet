package rpc_testdata

type Req struct {
	Uid            int32
	Id             int32
	State          int32
	State1         int32
	Value          int32
	String         string
	ValueMap       map[int32]bool
	ValueArray     [4]int32
	ValueSliceList [][2]int32

	MapMap     map[int32]map[int32]bool
	MapList    map[int32][4]int32
	MapSList   map[int32][]int32
	SListMap   []map[int32]bool
	SListList  [][4]int32
	SListSList [][]int32
	ListMap    [4]map[int32]bool
	ListList   [4][4]int32
	ListSList  [4][]int32

	PtrInt32 *int32
	PtrInt64 *int64
	PtrStr *string
	PtrSlice *[]int32
	PtrMap *map[int32]int32
	Nest *Embed
}

type Embed struct {
	Id int32
	String string
	PtrStr *string
	PtrEmbed *Embed
}

type Resp struct {
	PtrInt *int32
	T map[int32]*Embed
	PtrEmbed *Embed
}

