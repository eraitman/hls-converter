; Inno Setup 스크립트. Actions에서 dist\ (exe + ffmpeg/ffprobe) 를 묶어 설치 파일을 만든다.
; UTF-8 BOM 필수(한글). 버전은 ISCC /DAppVersion=... 으로 전달.
#ifndef AppVersion
  #define AppVersion "dev"
#endif

[Setup]
AppId={{3DFE1C2C-0B48-4652-B35A-6DE3D08ED970}
AppName=HLS 변환기
AppVersion={#AppVersion}
DefaultDirName={autopf}\HLS Converter
DefaultGroupName=HLS 변환기
DisableProgramGroupPage=yes
PrivilegesRequired=admin
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
OutputDir=..\out
OutputBaseFilename=hls-converter-setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
UninstallDisplayIcon={app}\hls-converter.exe

[Languages]
#if FileExists(CompilerPath + "Languages\Korean.isl")
Name: "korean"; MessagesFile: "compiler:Languages\Korean.isl"
#else
Name: "english"; MessagesFile: "compiler:Default.isl"
#endif

[Files]
Source: "..\dist\*"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\HLS 변환기"; Filename: "{app}\hls-converter.exe"
Name: "{group}\HLS 변환기 제거"; Filename: "{uninstallexe}"

[Run]
Filename: "{app}\hls-converter.exe"; Description: "HLS 변환기 실행"; Flags: nowait postinstall skipifsilent
