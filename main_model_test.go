package playlist_module

import (
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) {
	session := NewPlaylist()
	testID := 0
	fileBytes, err := os.ReadFile("./mp3_1.mp3") // for tests used one track with different duration
	if err != nil {
		panic("reading my-file.mp3 failed: " + err.Error())
	}
	t.Logf("\tTest %d:\tPlay test", testID)
	{
		song := Track{Name: "test", Duration: 5 * time.Second}
		session.AddSong(song)
		session.AddChunk(fileBytes)
		session.UnlockData()
		session.Play()
		if session.timer == nil || session.PlayingStatus() == false {
			t.Fatal("Timer is not created or Flag is false")
		}
		time.Sleep(time.Second * 6)
		if session.timer != nil || session.PlayingStatus() == true {
			t.Fatal("Timer is not expanse or Flag is true")
		}
	}
	testID++
	t.Logf("\tTest %d:\tTimer existence", testID)
	{
		i := 0
		session.AddChunk(fileBytes)
		session.UnlockData()
		session.Play()
		for session.PlayingStatus() && i < 110 {
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
		session.AddChunk(fileBytes)
		session.UnlockData()
		session.Play()
		go func() {
			for session.timer != nil && i < 130 {
				time.Sleep(time.Millisecond * 50)
				if session.PlayingStatus() == true {
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
	t.Logf("\tTest %d:\tNext test", testID) // And Prev
	{
		song := Track{Name: "test1", Duration: 1 * time.Second}
		session.AddSong(song)
		session.AddChunk(fileBytes)
		session.UnlockData()
		ch1 := make(chan bool)
		ch2 := make(chan bool)
		go func() {
			<-ch1
			<-ch2
		}()
		session.Play()
		go func() {
			time.Sleep(time.Millisecond * 200)
			session.AddChunk(fileBytes)
			session.UnlockData()
		}()
		if !session.Next(ch1) {
			t.Fatal("Next error")
		}
		if session.Current.Value.(Track).Name != "test1" {
			t.Fatal("Wrong track")
		}
		if session.Next(ch2) {
			t.Fatal("Next wrong usage")
		}
		time.Sleep(time.Second)
	}
	testID++
	t.Logf("\tTest %d:\tAdvanced Play test", testID)
	{
		song := Track{Name: "test2", Duration: 2 * time.Second}
		session.AddSong(song)
		song1 := Track{Name: "test3", Duration: 3 * time.Second}
		session.AddSong(song1)
		session.AddChunk(fileBytes)
		session.UnlockData()
		go func() {
			<-session.Ch
			session.AddChunk(fileBytes)
			session.UnlockData()
			<-session.Ch
			session.AddChunk(fileBytes)
			session.UnlockData()
		}()
		session.Play()
		time.Sleep(time.Second * 7)
		if session.Current.Next() != nil || session.PlayingStatus() {
			t.Fatal("Wrong position")
		}
	}
	testID++
	t.Logf("\tTest %d:\tDeleteSong test", testID)
	{
		session.DeleteSong(3)
		el := session.List.Front()
		if el.Value.(Track).Duration != 5*time.Second {
			t.Fatal("Wrong track deleted")
		}
		el = el.Next()
		if el.Value.(Track).Duration != 1*time.Second {
			t.Fatal("1-Wrong track deleted")
		}
		el = el.Next()
		if el.Value.(Track).Duration != 2*time.Second {
			t.Fatal("2-Wrong track deleted")
		}
		session.DeleteSong(1)
		el = session.List.Front()
		if el.Value.(Track).Duration != 5*time.Second {
			t.Fatal("3-Wrong track deleted")
		}
		el = el.Next()
		if el.Value.(Track).Duration != 2*time.Second {
			t.Fatal("4-Wrong track deleted")
		}
	}
	testID++
	t.Logf("\tTest %d:\tDeleteSong test", testID)
	{
		session.Destructor()
	}
}
