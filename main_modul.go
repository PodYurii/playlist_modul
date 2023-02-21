package playlist_modul

import (
	"container/list"
	"fmt"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
	"os"
	"sync"
)

type track struct {
	File     *os.File
	Name     string
	Duration uint
}

type Session struct {
	UserId  uint
	mutex   sync.Mutex
	List    *list.List
	current *list.Element
	context oto.Context
	player  oto.Player
}

func NewSession(id uint) *Session { // Constructor
	var newObj Session
	newObj.UserId = id
	newObj.List = list.New()
	newObj.current = newObj.List.Front()
	newObj.createContext()
	return &newObj
}

func (obj *Session) Pause() {
	if obj.player != nil {
		obj.player.Pause()
	}
}

func (obj *Session) Play() {
	if obj.player == nil {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
				obj.player = nil
			}
		}()
		obj.createPlayer(obj.current.Value.(track).File)
	} else {
		obj.player.Play()
	}
}

func (obj *Session) createPlayer(file *os.File) {
	decodedMp3, err := mp3.NewDecoder(file)
	if err != nil {
		panic("mp3.NewDecoder failed: " + err.Error())
	}
	obj.player = obj.context.NewPlayer(decodedMp3)
}

func (obj *Session) Prev() {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if obj.current.Prev() == nil {
		panic("This is a first song in list")
	}
	obj.current = obj.current.Prev()
}

func (obj *Session) Next() {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if obj.current.Next() == nil {
		panic("This is a last song in list")
	}
	obj.current = obj.current.Next()
}

func (obj *Session) AddSong(newTrack track) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	obj.List.PushBack(newTrack)
}

func (obj *Session) createContext() {
	otoCtx, readyChan, err := oto.NewContext(44100, 2, 2)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan
	obj.context = *otoCtx
}
