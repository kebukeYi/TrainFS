package service

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"sort"
	"strings"
	"testing"
	"trainfs/src/common"
)

func TestGo(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"", ""},
		{"/", ""},
		{"/user/local/app/hhh.png", "/user/local/app"},
		{"/hhh.png", ""},
		{"user/local/app/hhh.png", "user/local/app"},
	}

	for _, test := range tests {
		result := goFunc(test.path)
		if result != test.expected {
			t.Errorf("goFunc(%s) = %s; want %s", test.path, result, test.expected)
		}
	}
}

func goFunc(path string) string {
	if path == "/" || path == "" {
		return ""
	}

	index := strings.LastIndex(path, "/")
	path = path[0:index]

	if path == "" {
		return ""
	}

	return path
}

func TestForPath(t *testing.T) {
	//path := "user/local/app/hhh.png"
	path := "/user/local/app/"
	path1 := "/user/local/app"
	path2 := "/user"
	path3 := "/"
	split := strings.Split(path, "/")
	split1 := strings.Split(path1, "/")
	split2 := strings.Split(path2, "/")
	split3 := strings.Split(path3, "/")
	fmt.Println(split)       // [ user local app ]
	fmt.Println(split1)      // [ user local app]
	fmt.Println(split2)      // [ user]
	fmt.Println(split3)      // [ ]
	fmt.Println(len(split))  // 5
	fmt.Println(len(split1)) // 4
	fmt.Println(len(split2)) // 2
	fmt.Println(len(split3)) // 2
}

func TestSplitFileNamePath(t *testing.T) {
	tests := []struct {
		fileNamePath string
		expectedFile string
		expectedPath string
	}{
		{"", "", ""},
		{"/", "", "/"},
		{"/user/app", "app", "/user"},
		{"file.txt", "file.txt", ""},
		{"path/to/file.txt", "file.txt", "path/to"},
		{"/path/to/file.txt", "file.txt", "/path/to"},
		{"/file.txt", "file.txt", "/"},
		{"/app", "app", "/"},
	}

	for _, test := range tests {
		fileName, path := splitFileNamePath(test.fileNamePath)
		fmt.Printf("fileName:%s; path:%s \n", fileName, path)
		//if fileName != test.expectedFile || path != test.expectedPath {
		//	t.Errorf("splitFileNamePath(%s) = (%s, %s), expected (%s, %s)", test.fileNamePath, fileName, path, test.expectedFile, test.expectedPath)
		//}
	}
}

func TestBySortDataNodeFreeSpace(t *testing.T) {
	// Create test data
	node1 := &DataNodeInfo{FreeSpace: 100}
	node2 := &DataNodeInfo{FreeSpace: 200}
	node3 := &DataNodeInfo{FreeSpace: 50}
	nodes := []*DataNodeInfo{node1, node2, node3}

	// Sort nodes by free space
	sort.Sort(ByFreeSpace(nodes))

	for _, node := range nodes {
		fmt.Println(node.FreeSpace)
	}

	// Check if the nodes are sorted correctly
	if nodes[0].FreeSpace != 50 || nodes[1].FreeSpace != 100 || nodes[2].FreeSpace != 200 {
		//t.Errorf("Nodes are not sorted correctly")
	}
}

type MockLevelDB struct {
	fileMetaMap map[string]*FileMeta
}

func (m *MockLevelDB) GetFileMeta(path string) (*FileMeta, error) {
	fileMeta, ok := m.fileMetaMap[path]
	if !ok {
		return nil, leveldb.ErrNotFound
	}
	return fileMeta, nil
}

func (m *MockLevelDB) PutFileMeta(path string, fileMeta *FileMeta) error {
	m.fileMetaMap[path] = fileMeta
	return nil
}

func (m *MockLevelDB) SetFileMeta(path string, fileMeta *FileMeta) {
	m.fileMetaMap[path] = fileMeta
}

func (nn *MockLevelDB) checkPathOrCreate(path string, notCreate bool) (*FileMeta, error) {
	if path == "/" {
		fileMeta, err := nn.GetFileMeta(path)
		if err == leveldb.ErrNotFound && notCreate {
			fileMeta := &FileMeta{
				IsDir:     true,
				ChildList: make(map[string]*FileMeta),
				FileName:  path,
				FileSize:  0,
			}
			nn.PutFileMeta(path, fileMeta)
		}
		return fileMeta, err
	}
	rootFileMeta, _ := nn.GetFileMeta("/")
	split := strings.Split(path, "/")[1:]
	for i := 0; i < len(split); i++ {
		dir := split[i] // app
		file, ok := rootFileMeta.ChildList[dir]
		if !ok {
			if notCreate {
				fileMeta := &FileMeta{
					IsDir:     true,
					ChildList: make(map[string]*FileMeta),
					FileName:  CreatKeyFileName(rootFileMeta.FileName, dir),
					FileSize:  0,
				}
				nn.PutFileMeta(fileMeta.FileName, fileMeta)
				rootFileMeta.ChildList[dir] = fileMeta
				nn.PutFileMeta(rootFileMeta.FileName, rootFileMeta)
				rootFileMeta = fileMeta
				continue
			}
			return nil, common.ErrNotDir
		}
		rootFileMeta = file
	}
	return rootFileMeta, nil
}

func TestNameNodeCheckPathOrCreate(t *testing.T) {
	nn := &MockLevelDB{fileMetaMap: map[string]*FileMeta{}}
	fileMeta := &FileMeta{
		IsDir:     false,
		ChildList: make(map[string]*FileMeta),
		FileName:  "/",
		FileSize:  10,
	}
	nn.SetFileMeta("/", fileMeta)

	t.Run("Path exists", func(t *testing.T) {
		// Setup
		path := "/dir/file"
		fileMeta := &FileMeta{
			IsDir:     false,
			ChildList: make(map[string]*FileMeta),
			FileName:  path,
			FileSize:  10,
		}
		nn.SetFileMeta(path, fileMeta)

		// Execute
		result, _ := nn.checkPathOrCreate(path, false)
		fmt.Println(result)
		// Verify
		//assert.NoError(t, err)
		//assert.Equal(t, fileMeta, result)
	})

	t.Run("Path does not exist and notCreate is true", func(t *testing.T) {
		// Setup
		path := "/new/dir"

		// Execute
		result, _ := nn.checkPathOrCreate(path, true)
		fmt.Println(result)
		// Verify
		//assert.NoError(t, err)
		//assert.NotNil(t, result)
		//assert.True(t, result.IsDir)
		//assert.Equal(t, path, result.FileName)
	})

	t.Run("Path does not exist and notCreate is false", func(t *testing.T) {
		// Setup
		path := "/new/dir"
		nn.SetFileMeta("/", &FileMeta{
			IsDir:     true,
			ChildList: make(map[string]*FileMeta),
			FileName:  "/",
			FileSize:  0,
		})

		// Execute
		result, _ := nn.checkPathOrCreate(path, false)
		fmt.Println(result)
		// Verify
		//assert.Error(t, err)
		//assert.Equal(t, common.ErrNotDir, err)
		//assert.Nil(t, result)
	})
}
