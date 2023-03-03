package playlist_module

import (
	"container/list"
	"github.com/ivahaev/timer"
	"sync"
	"time"
)

type track struct {
	Name     string
	Duration time.Duration
	Id       uint64
}

type Playlist struct {
	List      *list.List
	Current   *list.Element
	IsPlaying bool
	timer     *timer.Timer
	mutex     sync.Mutex
}

func NewPlaylist() *Playlist {
	var obj Playlist
	obj.List = list.New()
	return &obj
}

func (obj *Playlist) Prev() bool {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if obj.Current.Prev() == nil {
		return false
	}
	obj.changeCurrent(obj.Current.Prev())
	return true
}

func (obj *Playlist) Next() bool {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if obj.Current.Next() == nil {
		return false
	}
	obj.changeCurrent(obj.Current.Next())
	return true
}

func (obj *Playlist) AddSong(newTrack track) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	obj.List.PushBack(newTrack)
	if obj.List.Len() == 1 {
		obj.Current = obj.List.Back()
	}
}

func (obj *Playlist) DeleteSong(pos int) bool {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	var el *list.Element
	if pos <= obj.List.Len()/2 {
		el = obj.List.Front()
		for i := 0; i < pos; i++ {
			el = el.Next()
		}
	} else {
		el = obj.List.Back()
		for i := obj.List.Len(); i > pos; i-- {
			el = el.Prev()
		}
	}
	if obj.Current == el {
		return false
	}
	obj.List.Remove(el)
	return true
}

func (obj *Playlist) Pause() {
	if obj.IsPlaying == true {
		obj.IsPlaying = false
		obj.timer.Pause()
	}
}

func (obj *Playlist) Play() {
	if obj.IsPlaying == false {
		if obj.timer == nil {
			if obj.Current != nil {
				obj.IsPlaying = true
				obj.timer = timer.AfterFunc(time.Second*obj.Current.Value.(track).Duration, func() {
					obj.timer = nil
					obj.IsPlaying = false
					if obj.Current.Next() != nil {
						obj.Current = obj.Current.Next()
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

func (obj *Playlist) changeCurrent(toChange *list.Element) {
	obj.timer.Stop()
	obj.timer = nil
	obj.IsPlaying = false
	obj.Current = toChange
	obj.Play()
}
