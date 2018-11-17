package session

import (
	"errors"
	// "fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-apibox/cache"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"gopkg.in/fsnotify.v1"
)

var cookieStores *cookieStoreCache
var cookieStoresMutex sync.RWMutex
var cookieStoreMutex sync.RWMutex

var defaultCookieStore *CookieStore

func init() {
	cookieStores = new(cookieStoreCache)
	cookieStores.stores = make(map[string]*CookieStore)
}

type CookieStore struct {
	sessions.CookieStore
	keyPairs     [][]byte
	SessionCache *cache.Cache
}

func DefaultCookieStore() (*CookieStore, error) {
	cookieStoreMutex.Lock()
	defer cookieStoreMutex.Unlock()

	if defaultCookieStore != nil {
		return defaultCookieStore, nil
	}

	var err error
	defaultCookieStore, err = NewCookieStore(true, "")
	return defaultCookieStore, err
}

func NewCookieStore(autoUpdate bool, keyPairsFile string) (*CookieStore, error) {
	var err error
	if keyPairsFile != "" {
		keyPairsFile, err = filepath.Abs(keyPairsFile)
		if err != nil {
			return nil, err
		}

		cookieStoreMutex.Lock()
		defer cookieStoreMutex.Unlock()

		if cookieStores.has(keyPairsFile) {
			return cookieStores.get(keyPairsFile), nil
		}
	}

	store := new(CookieStore)

	if keyPairsFile == "" {
		// 未指定文件，则自动生成
		store.GenerateKeyPairs()
		store.LoadKeyPairs(store.keyPairs)
	} else {
		// 指定了文件，则从文件中加载
		store.keyPairs, err = readKeyPairs(keyPairsFile)
		if err != nil {
			return nil, err
		}
		store.LoadKeyPairs(store.keyPairs)
	}

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   0, //session关闭即失效
		HttpOnly: true,
	}

	// session过期缓存，默认1小时过期，每分钟清理一次
	store.SessionCache = cache.NewCacheEx(time.Hour, time.Minute)

	// 每30天定期更新一次KEY
	if keyPairsFile != "" {
		// 更新KEY及文件
		go func() {
			for {
				// 每小时检测一次文件是否过期
				// time.Sleep(5 * time.Second)
				time.Sleep(time.Hour)

				err = store.updateKeyPairs(keyPairsFile)
				if err == nil {
					// 此处不需要更新，在检测文件变更的goroutine中会更新
					// store.Codecs = securecookie.CodecsFromPairs(keyPairs...)
				} else {
					log.Println("WARNING: key pairs auto update failed: " + err.Error())
				}
			}
		}()
	} else {
		go func() {
			for {
				// 此处实现key预生成，是为了避免因为KEY更新而其它cookiestore来不及导致cookie解码失败
				// 提前3天准备KEY
				time.Sleep(time.Duration(27*24) * time.Hour)
				store.GenerateKeyPairs()

				// 等3天后再实际加载KEY
				time.Sleep(time.Duration(3*24) * time.Hour)

				// 重新生成 key
				store.LoadKeyPairs(store.keyPairs)
			}
		}()
	}

	if keyPairsFile != "" {
		// 监控文件变化
		go store.watchKeyPairs(keyPairsFile, store.keyPairs)
		cookieStores.set(keyPairsFile, store)
	}

	return store, nil
}

func (store *CookieStore) GenerateKeyPairs() {
	genCount := 1
	if store.keyPairs == nil {
		store.keyPairs = make([][]byte, 4)
		genCount = 2
	}
	for i := 0; i < genCount; i++ {
		keyPair := make([][]byte, 0, 2)
		for i := 0; i < 2; i++ {
			key := securecookie.GenerateRandomKey(32)
			keyPair = append(keyPair, key)
		}
		store.keyPairs = append(keyPair, store.keyPairs[:2]...)
	}
}

func (store *CookieStore) LoadKeyPairs(keyPairs [][]byte) {
	store.Codecs = securecookie.CodecsFromPairs(keyPairs...)
}

func (store *CookieStore) GetKeyPairs() [][]byte {
	return store.keyPairs
}

func readKeyPairs(keyPairsFile string) ([][]byte, error) {
	// 检查文件是否存在，不存在则创建新的 keyParis
	var f *os.File
	var err error
	f, err = os.OpenFile(keyPairsFile, os.O_RDWR, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(keyPairsFile)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	bufLen := len(buf)
	if bufLen%32 != 0 {
		return nil, errors.New("Wrong key format in session key pairs file.")
	}

	var keyPairs [][]byte
	keyCount := int(bufLen / 32)
	if keyCount == 0 {
		// 生成新的 keyPair
		keyPairs = make([][]byte, 0, 2)
		for i := 0; i < 2; i++ {
			key := securecookie.GenerateRandomKey(32)
			keyPairs = append(keyPairs, key)
			f.Write(key)
		}
	} else {
		if keyCount%2 != 0 || keyCount > 4 {
			return nil, errors.New("Wrong key count in session key pairs file.")
		}
		keyPairs = make([][]byte, 0, keyCount)
		for i := 0; i < keyCount; i++ {
			keyPairs = append(keyPairs, buf[i*32:(i+1)*32])
		}
	}

	return keyPairs, nil
}

func (store *CookieStore) updateKeyPairs(keyPairsFile string) error {
	fInfo, err := os.Stat(keyPairsFile)
	if err != nil {
		return err
	}

	// 每30天重新生成，但提前3天预生成KEY
	// if time.Now().Before(fInfo.ModTime().Add(time.Duration(5) * time.Second)) {
	if time.Now().Before(fInfo.ModTime().Add(time.Duration(27*24) * time.Hour)) {
		return err
	}

	// 已过期，重新生成 key
	store.GenerateKeyPairs()

	go func() {
		// 等3天后再写入文件
		time.Sleep(time.Duration(3*24) * time.Hour)

		// 更新key文件
		f, err := os.Create(keyPairsFile)
		if err != nil {
			log.Println("Create key file failed:", err.Error())
			return
		}
		defer f.Close()

		for _, key := range store.keyPairs {
			f.Write(key)
		}
	}()

	return nil
}

func (store *CookieStore) delayReloadKeyPairs(changes chan bool, keyPairsFile string) {
	for {
		<-changes

		// 等待1秒后取出队列中的所有变更，避免频繁重载
		time.Sleep(time.Second)
		count := len(changes)
		for i := 0; i < count; i++ {
			<-changes
		}

		// // 调试：读取文件内容
		// f, _ := os.Open(keyPairsFile)
		// str, _ := ioutil.ReadAll(f)
		// fmt.Println(str)

		newKeyPairs, err := readKeyPairs(keyPairsFile)
		if err == nil {
			store.keyPairs = newKeyPairs
			store.LoadKeyPairs(store.keyPairs)
		} else {
			log.Println("WARNING: (session) reload key pairs failed: " + err.Error())
		}
	}
}

func (store *CookieStore) watchKeyPairs(keyPairsFile string, keyPairs [][]byte) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	changes := make(chan bool, 10)

	go store.delayReloadKeyPairs(changes, keyPairsFile)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// fmt.Println(event)

				if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					// 被删除或改名，则重新添加
					watcher.Remove(keyPairsFile)
					for {
						// 出错则间隔3秒重试
						err = watcher.Add(keyPairsFile)
						if err != nil {
							time.Sleep(3 * time.Second)
							continue
						}

						// 文件监控成功
						changes <- true
						break
					}
				}
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
					changes <- true
				}
			case err := <-watcher.Errors:
				log.Println("WARNING: (session) watcher error: " + err.Error())
			}
		}
	}()

	for {
		// 出错则间隔3秒重试
		err = watcher.Add(keyPairsFile)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}

	<-done
}

func (s *CookieStore) Get(r *http.Request, name string) (*Session, error) {
	session, err := s.CookieStore.Get(r, name)
	if err != nil {
		if session != nil {
			// cookie值不合法，清空
			session.Values = make(map[interface{}]interface{})
		} else {
			return nil, err
		}
	}

	var sessionId string
	newSession := false

	if sId, ok := session.Values["session_id"]; ok {
		sessionId, ok := sId.(string)
		if !ok {
			return nil, errors.New("Invalid session ID!")
		}
		_, found := s.SessionCache.Get(sessionId) // 更新TTL
		if !found {
			// session已过期，清除内容并重建
			session.Values = make(map[interface{}]interface{})
			newSession = true
		}
	} else {
		newSession = true
	}

	if newSession {
		// 新建session，id写入TTL缓存
		for {
			sessionId = RandSessionId()
			if s.SessionCache.SetIfNotExist(sessionId, true) {
				break
			}
		}
		session.Values["session_id"] = sessionId
	}

	return &Session{*session, r}, nil
}
