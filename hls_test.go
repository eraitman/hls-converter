package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPickRenditionsSkipsUpscale(t *testing.T) {
	got := pickRenditions(720)
	if len(got) != 2 || got[1].height != 720 {
		t.Fatalf("720p 원본은 480p,720p만 나와야 함: %+v", got)
	}
	if got := pickRenditions(360); len(got) != 1 || got[0].height != 360 {
		t.Fatalf("작은 원본은 원본 높이 1개: %+v", got)
	}
}

func TestBuildArgsNoAudio(t *testing.T) {
	args := strings.Join(buildArgs("in.mp4", "out", probeInfo{height: 1080}), " ")
	if strings.Contains(args, "0:a") || strings.Contains(args, "a:0") {
		t.Fatalf("오디오 없는 입력에 오디오 매핑이 들어감: %s", args)
	}
}

// 실제 ffmpeg로 짧은 영상을 변환해 본다. ffmpeg가 없으면 건너뜀.
func TestConvertEndToEnd(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg 없음")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "테스트 영상.mp4") // 한글+공백 경로
	gen := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "testsrc=size=1280x720:rate=30:duration=9",
		"-f", "lavfi", "-i", "sine=duration=9", "-shortest", "-pix_fmt", "yuv420p", src)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("테스트 영상 생성 실패: %v %s", err, out)
	}

	out, err := convert(context.Background(), src, func(float64) {})
	if err != nil {
		t.Fatal(err)
	}
	master, err := os.ReadFile(filepath.Join(out, "playlist.m3u8"))
	if err != nil {
		t.Fatal(err)
	}
	m := string(master)
	if !strings.Contains(m, "480p.m3u8") || !strings.Contains(m, "720p.m3u8") || strings.Contains(m, "1080p") {
		t.Fatalf("마스터 플레이리스트 이상:\n%s", m)
	}
	if _, err := os.Stat(filepath.Join(out, "720p_000.ts")); err != nil {
		t.Fatal(err)
	}
}
