param(
    [Parameter(Mandatory=$true)]
    [string]$Version
)

$ErrorActionPreference = "Stop"

# Validate version format (x.y.z)
if ($Version -notmatch '^\d+\.\d+\.\d+$') {
    Write-Error "Version must be in format x.y.z (e.g., 3.7.0)"
    exit 1
}

$parts = $Version -split '\.'
$major = $parts[0]
$minor = $parts[1]
$patch = $parts[2]
$build = "0"
$fullVersionWithBuild = "$Version.$build"

Write-Host "Updating to version: $Version" -ForegroundColor Green

function Set-FileContent($Path, $Content) {
    $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($Path, $Content, $utf8NoBom)
}

cd "${env:UserProfile}\git\link-router\"

# 1. cmd\linkrouter\main.go
$mainGoPath = "cmd\linkrouter\main.go"
$content = Get-Content $mainGoPath -Raw
$content = $content -replace 'version \d+\.\d+\.\d+', "version $Version"
Set-FileContent $mainGoPath $content

# 2. extension\updates.json
$updatesJsonPath = "extension\updates.json"
$content = Get-Content $updatesJsonPath -Raw
$content = $content -replace '"version": "\d+\.\d+\.\d+"', "`"version`": `"$Version`""
Set-FileContent $updatesJsonPath $content

# 3. extension\manifest.json
$manifestJsonPath = "extension\manifest.json"
$content = Get-Content $manifestJsonPath -Raw
$content = $content -replace '"version": "\d+\.\d+\.\d+"', "`"version`": `"$Version`""
Set-FileContent $manifestJsonPath $content

# 4. installer.iss
$installerPath = "installer.iss"
$content = Get-Content $installerPath -Raw
$content = $content -replace '(?<=#define MyAppVersion ").*(?=")', $Version
Set-FileContent $installerPath $content

# 5. cmd\linkrouter\versioninfo.json
$versionInfoPath1 = "cmd\linkrouter\versioninfo.json"
$content = Get-Content $versionInfoPath1 -Raw
$content = $content -replace '"FileVersion": "\d+\.\d+\.\d+\.0"', "`"FileVersion`": `"$fullVersionWithBuild`""
$content = $content -replace '"ProductVersion": "v\d+\.\d+\.\d+"', "`"ProductVersion`": `"v$Version`""
$content = $content -replace '"Major": \d+', "`"Major`": $major"
$content = $content -replace '"Minor": \d+', "`"Minor`": $minor"
$content = $content -replace '"Patch": \d+', "`"Patch`": $patch"
$content = $content -replace '"Build": \d+', "`"Build`": $build"
Set-FileContent $versionInfoPath1 $content

# 6. cmd\linkrouter-gui\versioninfo.json
$versionInfoPath2 = "cmd\linkrouter-gui\versioninfo.json"
$content = Get-Content $versionInfoPath2 -Raw
$content = $content -replace '"FileVersion": "\d+\.\d+\.\d+\.0"', "`"FileVersion`": `"$fullVersionWithBuild`""
$content = $content -replace '"ProductVersion": "v\d+\.\d+\.\d+"', "`"ProductVersion`": `"v$Version`""
$content = $content -replace '"Major": \d+', "`"Major`": $major"
$content = $content -replace '"Minor": \d+', "`"Minor`": $minor"
$content = $content -replace '"Patch": \d+', "`"Patch`": $patch"
$content = $content -replace '"Build": \d+', "`"Build`": $build"
Set-FileContent $versionInfoPath2 $content

Write-Host "Version updated successfully!" -ForegroundColor Green

# Build everything
go generate .\cmd\linkrouter\
go build -ldflags="-H windowsgui -s -w -buildid=" -trimpath -o bin\ .\cmd\linkrouter\
& "${env:ProgramFiles}\7-Zip\7z.exe" a -tzip "${env:UserProfile}\git\link-router\bin\linkrouter-extension.zip" "${env:UserProfile}\git\link-router\extension\*"
if ($LASTEXITCODE -eq 0) { 
    Copy-Item "${env:UserProfile}\git\link-router\bin\linkrouter-extension.zip" "${env:UserProfile}\git\link-router\bin\linkrouter-extension.xpi" 
}
go generate .\cmd\linkrouter-gui\
cd .\cmd\linkrouter-gui\
Wails build
cd ..\..
cp cmd\linkrouter-gui\build\bin\linkrouter-gui.exe bin\
cd .\cmd\linkrouter-gui\

Write-Host "Building Inno Setup installer..." -ForegroundColor Yellow
Start-Process -FilePath "${env:ProgramFiles(x86)}\Inno Setup 6\Compil32.exe" -ArgumentList "/cc", "${env:UserProfile}\git\link-router\installer.iss" -Wait -NoNewWindow

if ($LASTEXITCODE -eq 0) {
    Write-Host "Release $Version built successfully!" -ForegroundColor Green
    Write-Host "Output: bin\" -ForegroundColor Cyan
    Write-Host "Installer: bin\LinkRouter-Setup-$Version.exe" -ForegroundColor Cyan
} else {
    Write-Error "Inno Setup compilation failed!"  -ForegroundColor Red
    exit 1
}

