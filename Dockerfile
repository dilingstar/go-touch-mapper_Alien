FROM debian:13.1-slim
RUN apt-get update && apt-get install curl wget python3 -y \
    && rm -rf /var/lib/apt/lists/* \
    && curl -fsSL https://raw.githubusercontent.com/arduino/arduino-cli/master/install.sh | sh \
    && arduino-cli config init \
    && arduino-cli config add board_manager.additional_urls https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json \
    && arduino-cli core update-index \
    && arduino-cli core install esp32:esp32 \
    && arduino-cli cache clean \
    # arduino-cli compile --fqbn esp32:esp32:esp32s3 ./ -e
#docker run --rm  -v ./hid_touch:/hid_touch  espbuild arduino-cli compile --fqbn esp32:esp32:esp32s3 /hid_touch -e