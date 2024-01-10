FROM ubuntu:22.04

RUN apt-get update && apt-get dist-upgrade -y
RUN DEBIAN_FRONTEND=noninteractive \
    apt-get install -y --no-install-recommends nodejs npm git

# for phantomjs
RUN DEBIAN_FRONTEND=noninteractive \
    apt-get install -y --no-install-recommends libfontconfig

# for puppeteer
RUN DEBIAN_FRONTEND=noninteractive \
    apt-get install -y --no-install-recommends gconf-service libasound2 \
    libatk1.0-0 libcairo2 libcups2 libdbus-1-3 libdrm2 libgbm1 libgconf-2-4 \
    libgdk-pixbuf2.0-0 libglib2.0-0 libgtk-3-0 libnspr4 libpango-1.0-0 \
    libpangocairo-1.0-0 libx11-xcb1 libxcomposite1 libxcursor1 libxdamage1 \
    libxfixes3 libxi6 libxrandr2 libxrender1 libxss1 libxtst6 \
    fonts-liberation libappindicator1 libnss3 lsb-release xdg-utils wget

# for puppeteer, too
RUN adduser --system --group --shell /bin/sh heim
USER heim

# copy source code to /srv/heim/client/src
WORKDIR /srv/heim/client/

ENV PATH $PATH:node_modules/.bin

VOLUME /srv/heim/client/build
