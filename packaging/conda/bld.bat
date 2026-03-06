@echo off

:: Copy binaries to the conda environment Scripts directory
copy cws.exe %PREFIX%\Scripts\
copy prismd.exe %PREFIX%\Scripts\

:: Copy GUI if it exists
if exist prism-gui.exe (
    copy prism-gui.exe %PREFIX%\Scripts\
)

:: Add a message for users
echo Prism v0.4.1 has been installed.
echo To get started, run: cws test
echo.