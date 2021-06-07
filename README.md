# Purpose

"Registratura-bot" (see below for the meaning of this word).
This is a simple telegram bot I wrote for personal usage. It checks my status on a certain webpage and reports to me.

**Ethical disclaimer**: on some stage, this bot has to go behind a simple CAPTHCHA. I'm both hands against scraping
    WWW behind CAPTCHA's, and those are there for a reason. But in this particular case, neither scraping was happening
    nor was it for broad usage - only personal.

# Stack

Golang: bot itself, with chromedp under the hood to do automated page navigation
Python-OpenCV + PyTesseract: OCR part
Deployment: Docker

# Usage

I'm not expecing anybody to run it because it is *very* usecase-specific, but just a note for myself

```
docker build -t registratura-bot:testing . \
        && docker tag registratura-bot:testing <your_docker_id>/registratura-bot:testing \
        && docker push <your_docker_id>/registratura-bot:testing
```

deployment on a server (provide `config.yaml` under `PWD/configs`)
```
docker run -it -v `pwd`/configs:/app/configs -v `pwd`/imgs:/app/imgs --rm registratura-bot:testing
```

# Trivia

In Russian, "registratura" (регистратура) means a "front-desk" of some governmental place such as municipal
hospital, where appointments were set.
It became a meme that there was no notion of queue there and this place was typically overcrowded
in the mornings, with people aggressively trying to win their slot for doctor appointments for that day.
