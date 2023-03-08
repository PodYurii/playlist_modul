package playlist_module

import (
	"bytes"
	"container/list"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
	"github.com/ivahaev/timer"
	"log"
	"sync"
	"time"
)

type Track struct {
	Name     string
	Duration time.Duration
	Id       uint64
}

type Playlist struct {
	List    *list.List
	Current *list.Element
	Oto     *oto.Context
	player  oto.Player
	timer   *timer.Timer
	mutexL  sync.Mutex
	mutexD  sync.Mutex
	data    []byte
}

func NewPlaylist() *Playlist {
	var obj Playlist
	obj.List = list.New()
	otoCtx, readyChan, err := oto.NewContext(44100, 2, 2)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan
	obj.Oto = otoCtx
	obj.data = make([]byte, 0)
	obj.mutexD.Lock()
	return &obj
}

func (obj *Playlist) Destructor() {
	obj.mutexL.Lock()
	obj.Current = nil
	obj.PlayerClose()
	obj.List = nil
}

func (obj *Playlist) UnlockData() {
	obj.mutexD.Unlock()
}

func (obj *Playlist) ClearData() {
	obj.data = make([]byte, 0)
}

func (obj *Playlist) DataCheck() bool {
	if len(obj.data) == 0 {
		return true
	}
	return false
}

func (obj *Playlist) AddChunk(chunk []byte) {
	obj.data = append(obj.data, chunk...)
}

func (obj *Playlist) PlayerClose() {
	if obj.player != nil {
		obj.timer = nil
		obj.player.Pause()
		obj.data = make([]byte, 0)
		err := obj.player.Close()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (obj *Playlist) PlayingStatus() bool {
	if obj.player == nil {
		return false
	}
	return obj.player.IsPlaying()
}

func (obj *Playlist) Prev(ch chan bool) bool {
	obj.mutexL.Lock()
	defer obj.mutexL.Unlock()
	if obj.Current.Prev() == nil {
		ch <- false
		return false
	}
	obj.changeCurrent(obj.Current.Prev(), ch)
	return true
}

func (obj *Playlist) Next(ch chan bool) bool {
	obj.mutexL.Lock()
	defer obj.mutexL.Unlock()
	if obj.Current.Next() == nil {
		ch <- false
		return false
	}
	obj.changeCurrent(obj.Current.Next(), ch)
	return true
}

func (obj *Playlist) AddSong(newTrack Track) {
	obj.mutexL.Lock()
	defer obj.mutexL.Unlock()
	obj.List.PushBack(newTrack)
	if obj.List.Len() == 1 {
		obj.Current = obj.List.Back()
	}
}

func (obj *Playlist) DeleteSong(pos int) bool {
	obj.mutexL.Lock()
	defer obj.mutexL.Unlock()
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
		if obj.PlayingStatus() {
			return false
		}
		if obj.Current.Next() != nil {
			obj.Current = obj.Current.Next()
		} else if obj.Current.Prev() != nil {
			obj.Current = obj.Current.Prev()
		} else {
			obj.Current = nil
		}
		obj.PlayerClose()
	}
	obj.List.Remove(el)
	return true
}

func (obj *Playlist) Pause() {
	if obj.player.IsPlaying() == true {
		obj.player.Pause()
		obj.timer.Pause()
	}
}

func (obj *Playlist) Play() {
	if obj.PlayingStatus() == false {
		if obj.timer == nil {
			obj.mutexD.Lock()
			if len(obj.data) == 0 {
				log.Println("Empty data!")
				return
			}
			fileBytesReader := bytes.NewReader(obj.data)
			decodedMp3, err := mp3.NewDecoder(fileBytesReader)
			if err != nil {
				log.Println("mp3.NewDecoder failed: ", err)
				obj.data = make([]byte, 0)
				obj.mutexD.Unlock()
				return
			}
			obj.player = obj.Oto.NewPlayer(decodedMp3)
			obj.timer = timer.AfterFunc(obj.Current.Value.(Track).Duration, func() {
				obj.timer = nil
				obj.player.Pause()
				err := obj.player.Close()
				if err != nil {
					return
				}
				obj.data = make([]byte, 0)
				obj.mutexL.Lock()
				if obj.Current.Next() != nil {
					obj.Current = obj.Current.Next()
					obj.mutexL.Unlock()
					obj.Play()
					return
				}
				obj.mutexL.Unlock()
			})
			obj.timer.Start()
			obj.player.Play()
		} else {
			obj.player.Play()
			obj.timer.Start()
		}
	}
	return
}

func (obj *Playlist) changeCurrent(toChange *list.Element, ch chan bool) {
	if obj.player != nil {
		obj.PlayerClose()
	}
	obj.Current = toChange
	ch <- true
	obj.Play()
}
