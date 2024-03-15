package merkledag

/*
	常量声明块，定义了两个常量 FILE 和 DIR，它们的值分别是 0 和 1，使用了 iota 关键字，表示自增的枚举值
*/
const (
	FILE = iota
	DIR
)

/*
	规定了实现了 Node 接口的类型必须实现 Size() 方法和 Type() 方法。
	Size() 方法返回一个 uint64 类型的值，Type() 方法返回一个 int 类型的值
*/
type Node interface {
	Size() uint64
	Type() int
}

/*
	扩展了 Node 接口，并规定了实现了 File 接口的类型必须实现 Bytes() 方法。
	Bytes() 方法返回一个 []byte 类型的值，用于获取文件内容
*/
type File interface {
	Node

	Bytes() []byte
}

/*
	扩展了 Node 接口，并规定了实现了 Dir 接口的类型必须实现 It() 方法。
	It() 方法返回一个 DirIterator 类型的值，用于遍历文件夹中的文件和子文件夹
*/
type Dir interface {
	Node

	It() DirIterator
}

/*
	实现了 DirIterator 接口的类型必须实现 Next() 方法和 Node() 方法。
	Next() 方法用于检查是否还有下一个文件或文件夹，Node() 方法返回当前文件或文件夹
*/
type DirIterator interface {
	Next() bool

	Node() Node
}