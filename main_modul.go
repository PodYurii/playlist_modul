package playlist_module

import (
	"container/list"
	"github.com/ivahaev/timer"
	"sync"
	"time"
)

type Track struct {
	Name     string
	Duration time.Duration
	Id       uint64
}

type Playlist struct {
	List      *list.List
	Current   *list.Element
	isPlaying bool
	timer     *timer.Timer
	mutex     sync.Mutex
}

func NewPlaylist() *Playlist {
	var obj Playlist
	obj.List = list.New()
	return &obj
}

func (obj *Playlist) Destructor() {
	obj.mutex.Lock()
	obj.Current = nil
	if obj.timer != nil {
		obj.timer.Stop()
		obj.timer = nil
	}
	obj.List = nil
}

func (obj *Playlist) PlayingStatus() bool {
	return obj.isPlaying
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

func (obj *Playlist) AddSong(newTrack Track) {
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
	if pos > obj.List.Len()/2-1 {
		el = obj.List.Back()
		for i := obj.List.Len() - 1; i > pos; i-- {
			el = el.Prev()
		}
	} else {
		el = obj.List.Front()
		for i := 0; i < pos; i++ {
			el = el.Next()
		}
	}
	if obj.Current == el {
		if obj.isPlaying {
			return false
		}
		if obj.Current.Next() != nil {
			obj.Current = obj.Current.Next()
		} else if obj.Current.Prev() != nil {
			obj.Current = obj.Current.Prev()
		} else {
			obj.Current = nil
		}
	}
	obj.List.Remove(el)
	return true
}

func (obj *Playlist) Pause() {
	if obj.isPlaying == true {
		obj.isPlaying = false
		obj.timer.Pause()
	}
}

func (obj *Playlist) Play() bool {
	if obj.isPlaying == false {
		if obj.timer == nil {
			if obj.Current != nil {
				obj.isPlaying = true
				obj.timer = timer.AfterFunc(obj.Current.Value.(Track).Duration, func() {
					obj.timer = nil
					obj.isPlaying = false
					obj.mutex.Lock()
					if obj.Current.Next() != nil {
						obj.Current = obj.Current.Next()
						obj.mutex.Unlock()
						obj.Play()
						return
					}
					obj.mutex.Unlock()
				})
				obj.timer.Start()
			} else {
				return false
			}
		} else {
			obj.isPlaying = true
			obj.timer.Start()
		}
	}
	return true
}

func (obj *Playlist) changeCurrent(toChange *list.Element) {
	if obj.isPlaying == true {
		obj.timer.Stop()
		obj.timer = nil
		obj.isPlaying = false
	}
	obj.Current = toChange
	obj.Play()
}
