package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// 만들 화질 목록. 원본보다 큰 높이는 건너뛴다(업스케일 방지).
type rendition struct {
	height   int
	vBitrate int // kbps
	aBitrate int // kbps
}

var renditions = []rendition{
	{480, 1200, 96},
	{720, 2400, 128},
	{1080, 3600, 192},
}

const (
	segmentSeconds  = 4    // HLS 세그먼트 길이
	keyframeSeconds = 2    // 키프레임 간격 (segmentSeconds의 약수여야 함)
	maxrateRatio    = 1.07 // 비트레이트 허용 변동폭
	bufsizeRatio    = 1.5
)

type probeInfo struct {
	height   int
	duration float64 // 초
	hasAudio bool
}

func pickRenditions(srcHeight int) []rendition {
	var out []rendition
	for _, r := range renditions {
		if r.height <= srcHeight {
			out = append(out, r)
		}
	}
	if len(out) == 0 { // 원본이 가장 낮은 화질보다 작으면 원본 높이로 1개만
		r := renditions[0]
		r.height = srcHeight
		out = append(out, r)
	}
	return out
}

// 1080p 미만 화질에 NVENC를 쓰는 높이 기준. 1080p는 화질 우선으로 libx264.
const cpuMinHeight = 1080

// NVIDIA GPU + NVENC 사용 가능 여부를 짧은 테스트 인코딩으로 1회 확인한다.
var nvencAvailable = sync.OnceValue(func() bool {
	cmd := exec.Command(tool("ffmpeg"), "-v", "error", "-f", "lavfi", "-i", "color=s=256x256:d=0.1",
		"-c:v", "h264_nvenc", "-f", "null", "-")
	hideWindow(cmd)
	return cmd.Run() == nil
})

// gpu가 true면 1080p 미만은 h264_nvenc, 1080p는 libx264로 인코딩한다.
func buildArgs(input, outDir string, info probeInfo, gpu bool) []string {
	rs := pickRenditions(info.height)
	n := len(rs)

	filter := fmt.Sprintf("[0:v]split=%d", n)
	for i := range rs {
		filter += fmt.Sprintf("[s%d]", i)
	}
	for i, r := range rs {
		filter += fmt.Sprintf(";[s%d]scale=-2:%d[v%d]", i, r.height, i)
	}

	args := []string{"-hide_banner", "-y", "-nostats", "-progress", "pipe:1", "-i", input, "-filter_complex", filter}
	for i := range rs {
		args = append(args, "-map", fmt.Sprintf("[v%d]", i))
		if info.hasAudio {
			args = append(args, "-map", "0:a:0")
		}
	}

	args = append(args,
		"-profile:v", "high",
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", keyframeSeconds))
	for i, r := range rs {
		s := func(opt string) string { return fmt.Sprintf("-%s:v:%d", opt, i) }
		if gpu && r.height < cpuMinHeight {
			args = append(args, s("c"), "h264_nvenc", s("preset"), "p5", s("rc"), "vbr",
				s("spatial-aq"), "1", s("forced-idr"), "1")
		} else {
			args = append(args, s("c"), "libx264", s("preset"), "medium", s("sc_threshold"), "0")
		}
		args = append(args,
			fmt.Sprintf("-b:v:%d", i), fmt.Sprintf("%dk", r.vBitrate),
			fmt.Sprintf("-maxrate:v:%d", i), fmt.Sprintf("%dk", int(float64(r.vBitrate)*maxrateRatio)),
			fmt.Sprintf("-bufsize:v:%d", i), fmt.Sprintf("%dk", int(float64(r.vBitrate)*bufsizeRatio)))
	}

	if info.hasAudio {
		args = append(args, "-c:a", "aac", "-ar", "48000")
		for i, r := range rs {
			args = append(args, fmt.Sprintf("-b:a:%d", i), fmt.Sprintf("%dk", r.aBitrate))
		}
	}

	var streamMap []string
	for i, r := range rs {
		if info.hasAudio {
			streamMap = append(streamMap, fmt.Sprintf("v:%d,a:%d,name:%dp", i, i, r.height))
		} else {
			streamMap = append(streamMap, fmt.Sprintf("v:%d,name:%dp", i, r.height))
		}
	}

	return append(args,
		"-f", "hls", "-hls_time", strconv.Itoa(segmentSeconds), "-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-hls_segment_filename", filepath.Join(outDir, "%v_%03d.ts"),
		"-master_pl_name", "playlist.m3u8",
		"-var_stream_map", strings.Join(streamMap, " "),
		filepath.Join(outDir, "%v.m3u8"))
}

// PATH의 ffmpeg/ffprobe를 우선 쓰고, 없으면 exe와 같은 폴더에서 찾는다.
func tool(name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return name
}

func probe(ctx context.Context, input string) (probeInfo, error) {
	cmd := exec.CommandContext(ctx, tool("ffprobe"), "-v", "error",
		"-show_entries", "stream=codec_type,height:format=duration", "-of", "json", input)
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return probeInfo{}, fmt.Errorf("ffprobe 실패: %w", err)
	}
	var res struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Height    int    `json:"height"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		return probeInfo{}, err
	}
	var info probeInfo
	for _, s := range res.Streams {
		switch s.CodecType {
		case "video":
			if info.height == 0 {
				info.height = s.Height
			}
		case "audio":
			info.hasAudio = true
		}
	}
	if info.height == 0 {
		return info, fmt.Errorf("비디오 스트림이 없습니다")
	}
	info.duration, _ = strconv.ParseFloat(res.Format.Duration, 64)
	return info, nil
}

// input 옆에 확장자를 뺀 이름의 폴더를 만들고 HLS로 변환한다. onProgress는 0~1.
func convert(ctx context.Context, input string, onProgress func(float64)) (string, error) {
	info, err := probe(ctx, input)
	if err != nil {
		return "", err
	}
	outDir := strings.TrimSuffix(input, filepath.Ext(input))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, tool("ffmpeg"), buildArgs(input, outDir, info, nvencAvailable())...)
	hideWindow(cmd)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("ffmpeg 실행 실패: %w", err)
	}
	sc := bufio.NewScanner(stdout)
	for sc.Scan() {
		if v, ok := strings.CutPrefix(sc.Text(), "out_time_us="); ok && info.duration > 0 {
			if us, err := strconv.ParseFloat(v, 64); err == nil {
				onProgress(min(us/1e6/info.duration, 1))
			}
		}
	}
	if err := cmd.Wait(); err != nil {
		msg := stderr.String()
		if len(msg) > 500 {
			msg = msg[len(msg)-500:]
		}
		return "", fmt.Errorf("ffmpeg 오류: %v\n%s", err, msg)
	}
	return outDir, nil
}
