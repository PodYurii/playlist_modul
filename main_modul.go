package playlist_modul

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
	current   *list.Element
	IsPlaying bool
	timer     *timer.Timer
	mutex     sync.Mutex
}

func (obj *Playlist) Prev() bool {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if obj.current.Prev() == nil {
		return false
	}
	obj.current.Value = obj.current.Prev()
	return true
}

func (obj *Playlist) Next() bool {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	if obj.current.Next() == nil {
		return false
	}
	obj.current.Value = obj.current.Next()
	return true
}

func (obj *Playlist) AddSong(newTrack track) {
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	obj.List.PushBack(newTrack)
	if obj.List.Len() == 1 {
		obj.current = obj.List.Back()
	}
}

func (obj *Playlist) DeleteSong(Track string) bool {
	if obj.current.Value.(track).Name == Track {
		return false
	}
	obj.mutex.Lock()
	defer obj.mutex.Unlock()
	el := obj.List.Front()
	for i := 0; i < obj.List.Len(); i++ {
		if el.Value.(track).Name == Track {
			obj.List.Remove(el)
			return true
		}
		el = el.Next()
	}
	return false
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

func (obj *Playlist) changeCurrent(toChange *list.Element) {
	obj.timer.Stop()
	obj.timer = nil
	obj.IsPlaying = false
	obj.current = toChange
	obj.Play()
}
