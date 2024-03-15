package merkledag

import (
	"encoding/json"
	"hash"
)

// Object 结构体定义了一个对象，包含了 Links 数组和 Data 字节数组
type Object struct {
	Links []Link // 存储链接信息的数组
	Data  []byte // 存储数据的字节数组
}

// Link 结构体定义了一个链接，包含了名称、哈希值和大小信息
type Link struct {
	Name string // 文件或文件夹的名称
	Hash []byte // 文件或文件夹的哈希值
	Size int    // 文件或文件夹的大小
}

// Add 函数将节点数据分片并存储到 KVStore 中，并返回 Merkle 根
func Add(store KVStore, node Node, h hash.Hash) []byte {
	if node.Type() == FILE {
		// 如果节点类型为文件，则调用 sliceFile 函数对文件进行分片处理
		file := node.(File)
		fileSlice := sliceFile(file, store, h)
		jsonData, _ := json.Marshal(fileSlice)
		h.Write(jsonData)
		return h.Sum(nil)
	} else {
		// 如果节点类型为文件夹，则调用 sliceDirectory 函数对文件夹进行分片处理
		dir := node.(Dir)
		dirSlice := sliceDirectory(dir, store, h)
		jsonData, _ := json.Marshal(dirSlice)
		h.Write(jsonData)
		return h.Sum(nil)
	}
}

// dfsForSlice 函数对文件进行分片处理，并将分片数据存储到 KVStore 中
func dfsForSlice(hight int, node File, store KVStore, seedId int, h hash.Hash) (*Object, int) {
	if hight == 1 {
		// 如果高度为 1，表示当前文件已经是最底层的分片
		if (len(node.Bytes()) - seedId) <= 256*1024 {
			// 如果剩余数据不足 256KB，则将剩余数据作为一个分片
			data := node.Bytes()[seedId:]
			blob := Object{
				Links: nil,
				Data:  data,
			}
			jsonData, _ := json.Marshal(blob)
			h.Reset()
			h.Write(jsonData)
			exists, _ := store.Has(h.Sum(nil))
			if !exists {
				store.Put(h.Sum(nil), data)
			}
			return &blob, len(data)
		}
		// 否则，对剩余数据进行分片处理
		links := &Object{}
		totalLen := 0
		for i := 1; i <= 4096; i++ {
			end := seedId + 256*1024
			if len(node.Bytes()) < end {
				end = len(node.Bytes())
			}
			data := node.Bytes()[seedId:end]
			blob := Object{
				Links: nil,
				Data:  data,
			}
			totalLen += len(data)
			jsonData, _ := json.Marshal(blob)
			h.Reset()
			h.Write(jsonData)
			exists, _ := store.Has(h.Sum(nil))
			if !exists {
				store.Put(h.Sum(nil), data)
			}
			links.Links = append(links.Links, Link{
				Hash: h.Sum(nil),
				Size: len(data),
			})
			links.Data = append(links.Data, []byte("data")...)
			seedId += 256 * 1024
			if seedId >= len(node.Bytes()) {
				break
			}
		}
		jsonData, _ := json.Marshal(links)
		h.Reset()
		h.Write(jsonData)
		exists, _ := store.Has(h.Sum(nil))
		if !exists {
			store.Put(h.Sum(nil), jsonData)
		}
		return links, totalLen
	} else {
		// 如果高度大于 1，表示需要继续递归处理分片
		links := &Object{}
		totalLen := 0
		for i := 1; i <= 4096; i++ {
			if seedId >= len(node.Bytes()) {
				break
			}
			child, childLen := dfsForSlice(hight-1, node, store, seedId, h)
			totalLen += childLen
			jsonData, _ := json.Marshal(child)
			h.Reset()
			h.Write(jsonData)
			links.Links = append(links.Links, Link{
				Hash: h.Sum(nil),
				Size: childLen,
			})
			typeName := "link"
			if child.Links == nil {
				typeName = "data"
			}
			links.Data = append(links.Data, []byte(typeName)...)
		}
		jsonData, _ := json.Marshal(links)
		h.Reset()
		h.Write(jsonData)
		exists, _ := store.Has(h.Sum(nil))
		if !exists {
			store.Put(h.Sum(nil), jsonData)
		}
		return links, totalLen
	}
}

// sliceFile 函数对文件进行分片处理
func sliceFile(node File, store KVStore, h hash.Hash) *Object {
	if len(node.Bytes()) <= 256*1024 {
		// 如果文件大小不超过 256KB，则不进行分片处理
		data := node.Bytes()
		blob := Object{
			Links: nil,
			Data:  data,
		}
		jsonData, _ := json.Marshal(blob)
		h.Reset()
		h.Write(jsonData)
		exists, _ := store.Has(h.Sum(nil))
		if !exists {
			store.Put(h.Sum(nil), data)
		}
		return &blob
	}
	// 否则，根据文件大小进行分片处理
	linkLen := (len(node.Bytes()) + (256*1024 - 1)) / (256 * 1024)
	hight := 0
	tmp := linkLen
	for {
		hight++
		tmp /= 4096
		if tmp == 0 {
			break
		}
	}
	res, _ := dfsForSlice(hight, node, store, 0, h)
	return res
}

// sliceDirectory 函数对文件夹进行分片处理
func sliceDirectory(node Dir, store KVStore, h hash.Hash) *Object {
	iter := node.It()
	tree := &Object{}
	for iter.Next() {
		elem := iter.Node()
		if elem.Type() == FILE {
			// 如果是文件，则调用 sliceFile 函数进行分片处理
			file := elem.(File)
			fileSlice := sliceFile(file, store, h)
			jsonData, _ := json.Marshal(fileSlice)
			h.Reset()
			h.Write(jsonData)
			tree.Links = append(tree.Links, Link{
				Hash: h.Sum(nil),
				Size: int(file.Size()),
				Name: file.Name(),
			})
			elemType := "link"
			if fileSlice.Links == nil {
				elemType = "data"
			}
			tree.Data = append(tree.Data, []byte(elemType)...)
		} else {
			// 如果是文件夹，则递归调用 sliceDirectory 函数进行分片处理
			dir := elem.(Dir)
			dirSlice := sliceDirectory(dir, store, h)
			jsonData, _ := json.Marshal(dirSlice)
			h.Reset()
			h.Write(jsonData)
			tree.Links = append(tree.Links, Link{
				Hash: h.Sum(nil),
				Size: int(dir.Size()),
				Name: dir.Name(),
			})
			elemType := "tree"
			tree.Data = append(tree.Data, []byte(elemType)...)
		}
	}
	jsonData, _ := json.Marshal(tree)
	h.Reset()
	h.Write(jsonData)
	exists, _ := store.Has(h.Sum(nil))
	if !exists {
		store.Put(h.Sum(nil), jsonData)
	}
	return tree
}