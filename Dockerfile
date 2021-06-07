# Building
FROM golang:1.16 as builder
RUN mkdir /build

WORKDIR /build
COPY . .

RUN go get -d ./...
RUN go build -o bot-server
# finished building


FROM python:3.7-slim

RUN apt update && apt install -y \
    tesseract-ocr \
    chromium \
    && rm -rf /var/lib/apt/lists/*

RUN pip install opencv-python \
    pytesseract \
    numpy

RUN mkdir /app

WORKDIR /app
COPY --from=builder /build/bot-server .
COPY ocr.py .

# executable
CMD ["./bot-server"]
