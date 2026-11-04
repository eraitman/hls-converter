---
name: owner-decisions
description: 오너가 직접 내린 선택과 이유 — 인코더 혼합, ffmpeg 우선순위, 설치 방식, 커밋 타이밍
metadata:
  type: project
---

- **인코더:** 처음엔 화질 때문에 전부 libx264 → 속도 때문에 혼합으로 변경. 480p·720p는 NVENC(저화질이라 무방), 1080p는 libx264. 1080p를 NVENC로 바꾸자고 제안하지 말 것.
  1080p 스트림 복사(-c:v copy) 아이디어는 오너가 거절함.
- **병렬 파일 처리:** 검토 결과 x264가 CPU를 이미 다 써서 이득 없음 → 안 함.
- **ffmpeg 우선순위:** PATH 우선, 없으면 exe 옆 동봉본 (오너 요청).
- **설치 파일:** Inno Setup, Program Files, 시작 메뉴만. 바탕화면 바로가기 넣지 말 것.
- **커밋·푸시:** 기능 수정마다 하지 말고 개발 완료 후 한 번에.

**Why:** 오너가 대화에서 명시적으로 고른 것들이라 코드만 봐서는 "왜"가 안 보인다.
**How to apply:** 이 전제를 유지하고, 바꾸려면 먼저 오너에게 묻는다.

## 미검증 (2026-11-04 기준)
- NVENC 혼합 경로: Mac/CI에 GPU가 없어 오너 Windows PC에서만 확인 가능. 아직 결과 못 받음.
- Inno Setup 설치 파일 빌드: CI에서 첫 실행 결과 확인 전.
- Actions 실행 e8f2113, fce101a가 "cancelled"로 기록됨 — 원인 미확인.

관련: [[user-workflow]]
