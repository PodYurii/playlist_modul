package playlist_modul

import (
	"container/list"
	"github.com/ivahaev/timer"
	"os"
	"sync"
	"time"
)

type track struct {
	File     *os.File
	Name     string
	Duration time.Duration
}

type Session struct {
	UserId    uint
	mutex     sync.Mutex
	List      *list.List
	current   *list.Element
	IsPlaying bool
	timer     *timer.Timer
}

func NewSession(id uint) *Session { // Constructor
	var newObj Session
	newObj.UserId = id
	newObj.List = list.New()
	return &newObj
}

func (obj *Session) Pause() {
	if obj.IsPlaying == true {
		obj.IsPlaying = false
		obj.timer.Pause()
	}
}

func (obj *Session) Play() {
	if obj.IsPlaying == false {
		if obj.timer == nil {
			if obj.current != nil {
				obj.IsPlaying = true
				obj.timer = timer.AfterFunc(time.Second*obj.current.Value.(track).Duration, func() {
					obj.timer = nil
					obj.IsPlaying = false
					if obj.current.Next() != nil {
						obj.current = obj.current.Next()
						obj.Play()
					}
				})
				obj.timer.Start()
			} else {
				panic("List is empty")
			}
		} else {
			obj.IsPlaying = true
			obj.timer.Start()
		}
	}
}

func (obj *Session) Prev() {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if obj.current.Prev() == nil {
		panic("This is a first song in list")
	}
	obj.changeCurrent(obj.current.Prev())
}

func (obj *Session) Next() {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if obj.current.Next() == nil {
		panic("This is a last song in list")
	}
	obj.changeCurrent(obj.current.Next())
}

func (obj *Session) changeCurrent(toChange *list.Element) {
	obj.timer.Stop()
	obj.timer = nil
	obj.IsPlaying = false
	obj.current = toChange
	obj.Play()
}

func (obj *Session) AddSong(newTrack track) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	obj.List.PushBack(newTrack)
	if obj.List.Len() == 1 {
		obj.current = obj.List.Back()
	}
}
