package playlist_module

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	session := NewPlaylist()
	testID := 0
	t.Logf("\tTest %d:\tPlay test", testID)
	{
		song := Track{Name: "test", Duration: 5}
		session.AddSong(song)
		session.Play()
		if session.timer == nil || session.isPlaying == false {
			t.Fatal("Timer is not created or Flag is false")
		}
		time.Sleep(time.Second * 6)
		if session.timer != nil || session.isPlaying == true {
			t.Fatal("Timer is not expanse or Flag is true")
		}
	}
	testID++
	t.Logf("\tTest %d:\tTimer existence", testID)
	{
		i := 0
		session.Play()
		for session.isPlaying && i < 110 {
			time.Sleep(time.Millisecond * 50)
			i++
		}
		t.Logf("playing... %d * 50 millisec", i)
		if i < 95 || i > 100 {
			t.Fatal("Wrong play time")
		}
	}
	testID++
	t.Logf("\tTest %d:\tPause test", testID)
	{
		i := 0
		c := make(chan bool)
		session.Play()
		go func() {
			for session.timer != nil && i < 130 {
				time.Sleep(time.Millisecond * 50)
				if session.isPlaying == true {
					i++
				}
			}
			c <- true
		}()
		time.Sleep(time.Second * 2)
		session.Pause()
		time.Sleep(time.Second)
		session.Play()
		<-c
		t.Logf("playing... %d * 50 millisec", i)
		if i < 95 || i > 100 {
			t.Fatal("Wrong play time")
		}
	}
	testID++
	t.Logf("\tTest %d:\tNext test", testID)
	{
		song := Track{Name: "test1", Duration: 1}
		session.AddSong(song)
		session.Play()
		session.Next()
		if session.Current.Value.(Track).Name != "test1" {
			t.Fatal("Wrong track")
		}
	}
	testID++
	t.Logf("\tTest %d:\tAdvanced Play test", testID)
	{
		song := Track{Name: "test2", Duration: 2}
		session.AddSong(song)
		session.AddSong(song)
		session.Play()
		time.Sleep(time.Second * 6)
		if session.Current.Next() != nil {
			t.Fatal("Wrong position")
		}
	}
	testID++
	t.Logf("\tTest %d:\tDeleteSong test", testID)
	{
		session.DeleteSong(2)
		el := session.List.Front()
		if el.Value.(Track).Duration != 5 {
			t.Fatal("Wrong track deleted")
		}
		el = el.Next()
		if el.Value.(Track).Duration != 2 {
			t.Fatal("1-Wrong track deleted")
		}
		el = el.Next()
		if el.Value.(Track).Duration != 2 {
			t.Fatal("2-Wrong track deleted")
		}
		session.DeleteSong(2)
		el = session.List.Front()
		if el.Value.(Track).Duration != 5 {
			t.Fatal("3-Wrong track deleted")
		}
		el = el.Next()
		if el.Value.(Track).Duration != 2 {
			t.Fatal("4-Wrong track deleted")
		}
	}
	session.Destructor()
}
