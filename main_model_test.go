package playlist_modul

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	session := NewSession(0)
	testID := 0
	t.Logf("\tTest %d:\tPlay test", testID)
	{
		song := track{nil, "test", 5}
		session.AddSong(song)
		session.Play()
		if session.timer == nil || session.IsPlaying == false {
			t.Fatal("Timer is not created or Flag is false")
		}
		time.Sleep(time.Second * 6)
		if session.timer != nil || session.IsPlaying == true {
			t.Fatal("Timer is not expanse or Flag is true")
		}
	}
	testID++
	t.Logf("\tTest %d:\tTimer existence", testID)
	{
		i := 0
		session.Play()
		for session.IsPlaying && i < 110 {
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
				if session.IsPlaying == true {
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
		song := track{nil, "test1", 1}
		session.AddSong(song)
		session.Play()
		session.Next()
		if session.current.Value.(track).Name != "test1" {
			t.Fatal("Wrong track")
		}
	}
	testID++
	t.Logf("\tTest %d:\tAdvanced Play test", testID)
	{
		song := track{nil, "test2", 2}
		session.AddSong(song)
		session.AddSong(song)
		session.Play()
		time.Sleep(time.Second * 11)
		if session.current.Next() != nil {
			t.Fatal("Wrong position")
		}
	}
}
