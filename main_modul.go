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

type Track struct { // Element of List, containing all information about track
	Name     string
	Duration time.Duration
	Id       uint64 // Id on server side
}

type Playlist struct {
	List    *list.List
	Current *list.Element
	Oto     *oto.Context // main OtoV2 module for creating players of tracks(only one)
	player  oto.Player
	timer   *timer.Timer // timer with pause/stop
	mutexL  sync.Mutex   // mutex for List
	mutexD  sync.Mutex   // mutex for catching signal about track data ready state(like a self-refreshing channel) and 99% time is locked
	data    []byte       // track in bytes
	Ch      chan bool    // chan for downloading data for playing without stopping
}

func NewPlaylist() *Playlist { //Constructor
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
	obj.Ch = make(chan bool)
	return &obj
}

func (obj *Playlist) Destructor() {
	obj.mutexL.Lock()
	obj.Current = nil
	obj.PlayerClose()
	obj.List = nil
	obj.Ch <- false
	close(obj.Ch)
}

func (obj *Playlist) UnlockData() { //Called to inform playlist about readiness to play(after all data is written)
	obj.mutexD.Unlock()
}

func (obj *Playlist) ClearData() { //Called to clear slice after errors
	obj.data = make([]byte, 0)
}

func (obj *Playlist) DataCheck() bool { // Called to check data availability(it's a first play call for this track or no?)
	if len(obj.data) == 0 {
		return true
	}
	return false
}

func (obj *Playlist) AddChunk(chunk []byte) { // Called to write chunk of bytes after receiving another package from stream
	obj.data = append(obj.data, chunk...)
}

func (obj *Playlist) PlayerClose() { // Called to Close current playing track( after next, prev, delete, destructor)
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
	if pos > obj.List.Len()/2-1 { //counting from the nearest end
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
	if obj.Current == el { // if track is current, but is not played, we should change current pointer
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
		obj.PlayerClose() // stop timer and player for current track
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

func (obj *Playlist) Play() { //Current != nil check must take place before call(in GUI or interface)!!
	if obj.PlayingStatus() == false { // if track is playing just return
		if obj.timer == nil { // if we already have timer and player, we can just resume it
			obj.mutexD.Lock()       // wait for UnlockData call
			if len(obj.data) == 0 { // interface or stream error protection
				log.Println("Empty data!")
				return
			}
			fileBytesReader := bytes.NewReader(obj.data) // io.Reader for passing across bytes
			decodedMp3, err := mp3.NewDecoder(fileBytesReader)
			if err != nil {
				log.Println("mp3.NewDecoder failed: ", err)
				obj.data = make([]byte, 0)
				return
			}
			obj.player = obj.Oto.NewPlayer(decodedMp3)
			obj.timer = timer.AfterFunc(obj.Current.Value.(Track).Duration, func() { // func for switching to next track after ending of current
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
					obj.Ch <- true
					obj.Play()
					return
				}
				obj.mutexL.Unlock()
			})
			obj.player.Play()
			obj.timer.Start()
		} else {
			obj.player.Play()
			obj.timer.Start()
		}
	}
	return
}

func (obj *Playlist) changeCurrent(toChange *list.Element, ch chan bool) { // Called in Next and Prev(common code)
	if obj.player != nil {
		obj.PlayerClose()
	}
	obj.Current = toChange
	ch <- true // used to tell interface about readiness to download new data after changing current
	obj.Play()
}
