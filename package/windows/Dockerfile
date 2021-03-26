ARG SERVERCORE_VERSION
FROM mcr.microsoft.com/windows/servercore:${SERVERCORE_VERSION}
SHELL ["powershell", "-NoLogo", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]
COPY bin/drone-docker-image-digests.exe 'C:\drone-docker-image-digests\drone-docker-image-digests.exe'
CMD ["C:\drone-docker-image-digests\drone-docker-image-digests.exe"]
