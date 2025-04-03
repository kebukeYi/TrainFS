package service

import (
	"errors"
	"flag"
	"fmt"
	DBcommon "github.com/kebukeYi/TrainDB/common"
	"github.com/kebukeYi/TrainFS/common"
	"sort"
	"strings"
	"testing"
)

func TestFilePathSplit(t *testing.T) {
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
		result := SplitPath(test.path)
		if result != test.expected {
			t.Errorf("SplitPath(%s) = %s; want %s", test.path, result, test.expected)
		}
	}
}

func SplitPath(path string) string {
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
	fmt.Println(split)       // [ ,user,local,app, ]
	fmt.Println(split1)      // [ ,user,local,app]
	fmt.Println(split2)      // [ ,user]
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
		//{"file.txt", "file.txt", ""}, // error input
		{"path/to/file.txt", "file.txt", "path/to"},
		{"/path/to/file.txt", "file.txt", "/path/to"},
		{"/file.txt", "file.txt", "/"},
		{"/app", "app", "/"},

		{"\\root\\app\\y.jpg", "y.jpg", "\\root\\app"},

		//{"", "", ""},
		//{"\\", "", "\\"},
		//{"\\user\\app", "app", "\\user"},
		////{"file.txt", "file.txt", ""}, // error input
		//{"path\\to\\file.txt", "file.txt", "path\\to"},
		//{"\\path\\to\\file.txt", "file.txt", "\\path\\to"},
		//{"\\file.txt", "file.txt", "\\"},
		//{"\\app", "app", "\\"},
	}

	for _, test := range tests {
		fileName, path := common.SplitFileNamePath(test.fileNamePath)
		fmt.Printf("fileName:%s;expectedFile:%s; path:%s; expectedPath:%s \n",
			fileName, test.expectedFile, path, test.expectedPath)
		//if fileName != test.expectedFile || path != test.expectedPath {
		//	t.Errorf("splitFileNamePath(%s) = (%s, %s), expected (%s, %s)",
		//		test.fileNamePath, fileName, path, test.expectedFile, test.expectedPath)
		//}
	}
}

func TestNewNameNode(t *testing.T) {
	configFile := flag.String("conf", "./conf/nameNode_config.yml", "Path to conf file")
	flag.Parse()
	nameNode := &NameNode{}
	nameNode.Config = GetDataNodeConfig(configFile)
	fmt.Println(nameNode.Config)
}

func TestGetChunkIdOfFileChunkName_Success(t *testing.T) {
	tests := []struct {
		fileChunkName string
		expected      int32
	}{
		{"file_chunk_1", 1},
		{"file_chunk_2", 2},
		{"file", -1},
		{"", -1},
		{"/roo_t/app/t_est.txt_chunk_0", 0},
		{"/root/app/test.txt_chunk_0", 0},
		{"/root/app/test.txt_chunk_777", 777},
	}

	for _, test := range tests {
		result := common.GetChunkIdOfFileChunkName(test.fileChunkName)
		fmt.Println(result)
		if result != test.expected {
			t.Errorf("GetFileNameFromChunkName(%s) = %d; want %d",
				test.fileChunkName, result, test.expected)
		}
	}
}

func TestGetFileNameFromChunkName(t *testing.T) {
	tests := []struct {
		fileChunkName string
		expected      string
	}{
		{"file_chunk_1", "file"},
		{"file_chunk_2", "file"},
		{"file", ""},
		{"", ""},
		{"/roo_t/app/t_est.txt_chunk_0", "/roo_t/app/t_est.txt"},
		{"/root/app/test.txt_chunk_0", "/root/app/test.txt"},
		{"/root/app/test.txt_chunk_777", "/root/app/test.txt"},
	}

	for _, test := range tests {
		result := common.GetFileNameFromChunkName(test.fileChunkName)
		fmt.Println(result)
		if result != test.expected {
			t.Errorf("GetFileNameFromChunkName(%s) = %s; want %s", test.fileChunkName, result, test.expected)
		}
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
		return nil, DBcommon.ErrKeyNotFound
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

func (m *MockLevelDB) checkPathOrCreate(path string, notCreate bool) (*FileMeta, error) {
	if path == "/" {
		fileMeta, err := m.GetFileMeta(path)
		if errors.Is(err, DBcommon.ErrKeyNotFound) && notCreate {
			fileMeta := &FileMeta{
				IsDir:     true,
				ChildList: make(map[string]*FileMeta),
				FileName:  path,
				FileSize:  0,
			}
			m.PutFileMeta(path, fileMeta)
		}
		return fileMeta, err
	}
	rootFileMeta, _ := m.GetFileMeta("/")
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
				m.PutFileMeta(fileMeta.FileName, fileMeta)
				rootFileMeta.ChildList[dir] = fileMeta
				m.PutFileMeta(rootFileMeta.FileName, rootFileMeta)
				rootFileMeta = fileMeta
				continue
			}
			return nil, common.ErrNotDir
		}
		rootFileMeta = file
	}
	return rootFileMeta, nil
}

func TestGetOutBoundIP(t *testing.T) {
	if ip, err := common.GetOutBoundIP(); err != nil {
		t.Error(err)
	} else {
		fmt.Printf("ip: %s\n", ip)
	}
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
		result, err := nn.checkPathOrCreate(path, false)
		fmt.Printf("result:%v; err:%v\n", result, err)
	})

	t.Run("Path does not exist and notCreate is true", func(t *testing.T) {
		// Setup
		path := "/new/dir"

		// Execute
		result, err := nn.checkPathOrCreate(path, true)
		fmt.Printf("result:%v; err:%v\n", result, err)
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
		result, err := nn.checkPathOrCreate(path, false)
		fmt.Printf("result:%v; err:%v\n", result, err)
		// Verify
		//assert.Error(t, err)
		//assert.Equal(t, common.ErrNotDir, err)
		//assert.Nil(t, result)
	})
}
