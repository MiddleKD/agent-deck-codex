# Installing agent-deck-codex

이 저장소는 [asheshgoplani/agent-deck](https://github.com/asheshgoplani/agent-deck)의 포크입니다.
공식 릴리스 바이너리가 없으므로 소스에서 빌드합니다. **`install.sh`는 사용하지 마세요.**

## 사전 요구사항

- **Go 1.24.0** (정확히 — 1.25+는 TUI 렌더링 문제 있음)
- tmux, make

## 빌드 및 설치

기존 `agent-deck`과 공존하기 위해 모든 명령에 `BINARY_NAME`을 지정합니다:

```bash
git clone https://github.com/MiddleKD/agent-deck-codex.git
cd agent-deck-codex
make install-user BINARY_NAME=agent-deck-ex
# → ~/.local/bin/agent-deck-ex 로 설치됨
```

`~/.local/bin`이 PATH에 없으면 셸 설정 파일에 추가:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## 업데이트

```bash
git pull && make install-user BINARY_NAME=agent-deck-ex
```

## 제거

```bash
make uninstall-user BINARY_NAME=agent-deck-ex
```
