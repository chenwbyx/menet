//go:build ignore

//go:generate syncmap -name=MapInt32Int32 -pkg=protocol -o=./mapint32int32.go map[int32]int32
//go:generate syncmap -name=MapInt32MapInt32Int32 -pkg=protocol -o=./mapint32mapint32int32.go map[int32]*MapInt32Int32

package protocol
