# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **시작 전에 `.claude/memory/MEMORY.md`를 읽을 것.**
> "마무리하자"라고 하면 그 메모리를 갱신하고 커밋한다.

## 프로젝트 개요

여러 mp4를 끌어다 놓으면 순차적으로 HLS(480p/720p/1080p)로 변환하는 Windows용 GUI 도구.
Go + Fyne. exe는 GitHub Actions(windows-latest)에서 빌드하고, ffmpeg.exe/ffprobe.exe(BtbN GPL 빌드)를 동봉한다.
원래 `hls.sh`(bash로 ffmpeg 명령을 출력해 PowerShell에 붙여넣던 방식)를 대체한다.

## 개발 명령어

```bash
go test -v ./...                     # 단위 + 실제 ffmpeg 변환 테스트(ffmpeg 없으면 e2e는 skip)
go test -run TestConvertEndToEnd -v  # 단일 테스트
go run .                             # Mac에서 GUI 실행(로컬 ffmpeg/ffprobe는 PATH 사용)
git tag v0.1.0 && git push --tags    # Actions가 zip을 만들어 GitHub Release에 올림
```

main 브랜치 push만 해도 Actions artifact(hls-converter-win64)로 exe를 받을 수 있다.

## 아키텍처

- `hls.go`: ffprobe로 높이/길이/오디오 유무 조회 → ffmpeg **1회 실행**으로 `split` + `-var_stream_map` 사용해 전 화질 동시 인코딩, 마스터 플레이리스트(`playlist.m3u8`)도 ffmpeg가 생성. 원본보다 큰 화질은 건너뜀. 출력은 입력 파일 옆 `<파일명>/` 폴더.
- 키프레임은 `-force_key_frames`로 2초 고정(fps 파싱 불필요). `keyframeSeconds`는 `segmentSeconds`의 약수여야 세그먼트가 정렬된다.
- ffmpeg 경로: PATH 우선, 없으면 exe와 같은 폴더 (`tool()`).
- `proc_windows.go`: GUI exe(`-H=windowsgui`)에서 ffmpeg 콘솔 창이 뜨지 않게 CREATE_NO_WINDOW.
- Fyne UI 갱신은 고루틴에서 반드시 `fyne.Do`로.

## 반복 실수 주의

- Fyne는 CGO가 필요하다. Windows 빌드는 Actions 러너의 gcc(MinGW)에 의존한다.
- 인코더 혼합: NVENC가 되면 1080p 미만은 h264_nvenc, 1080p는 libx264(오너가 화질 때문에 1080p는 x264로 결정). 안 되면 전부 x264. 인코더별 옵션은 `-opt:v:N` 스트림 지정자로만 준다 — `-sc_threshold`는 x264 전용, `-crf`는 NVENC에 없음.
- NVENC 경로는 Mac/CI(GPU 없음)에서 검증 불가. 오너 Windows PC에서만 확인된다.
