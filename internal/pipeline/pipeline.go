package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"time"

	"github.com/agronskiy/telegram-page-checker-bot/internal/config"
	"github.com/agronskiy/telegram-page-checker-bot/internal/pipres"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"

	log "github.com/sirupsen/logrus"
)

func saveCaptchaImg(singleUrl *config.SingleURL, htmlIds *config.ElementIds, res *[]byte) chromedp.Tasks {
	log := log.WithField("name", singleUrl.Name)
	return chromedp.Tasks{
		emulation.SetDeviceMetricsOverride(1680, 1050, 1.0, false),
		chromedp.ActionFunc(func(context.Context) error {
			log.WithField("url", singleUrl.Url).Printf("Navigating to URL")
			return nil
		}),
		chromedp.Navigate(singleUrl.Url),
		chromedp.WaitVisible(htmlIds.CaptchaID, chromedp.ByID),
		chromedp.ActionFunc(func(context.Context) error {
			log.Printf("Saving CAPTCHA to file")
			return nil
		}),
		chromedp.Screenshot(htmlIds.CaptchaID, res, chromedp.NodeVisible),
	}
}

func submitDecodedCaptcha(
	singleUrl *config.SingleURL,
	htmlIds *config.ElementIds,
	captcha *[]byte, ok *bool,
) chromedp.Tasks {
	log := log.WithField("name", singleUrl.Name)
	return chromedp.Tasks{
		chromedp.SendKeys(htmlIds.CaptchaInputID, string(*captcha), chromedp.ByID),
		chromedp.Click(htmlIds.CaptchaButtonID, chromedp.ByID),
		chromedp.ActionFunc(func(context.Context) error {
			log.Println("Waiting after clicked CAPTCHA button")
			return nil
		}),
		chromedp.Sleep(1 * time.Second),
		chromedp.WaitVisible("footer", chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node
			if err := chromedp.Nodes(
				htmlIds.CaptchaErrID, &nodes, chromedp.AtLeast(0)).Do(ctx); err != nil {
				return err
			}
			if len(nodes) == 0 {
				*ok = true
			} else {
				*ok = false
			}
			return nil
		}),
	}
}

func proceedToSecondStage(
	singleUrl *config.SingleURL,
	htmlIds *config.ElementIds,
	result *pipres.PipelineResult,
) chromedp.Tasks {
	// There are two types of second stage.
	if singleUrl.Type == "initial" {
		return chromedp.Tasks{
			chromedp.WaitVisible("footer", chromedp.ByID),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var nodes []*cdp.Node
				if err := chromedp.Nodes(
					htmlIds.SecondStageButtonID, &nodes, chromedp.ByID, chromedp.AtLeast(0)).Do(ctx); err != nil {
					return err
				}
				if len(nodes) == 0 {
					*result = pipres.MaybeAlreadySigned
				}
				return nil
			}),
		}
	}

	return chromedp.Tasks{
		chromedp.WaitVisible("footer", chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node
			if err := chromedp.Nodes(
				htmlIds.SecondStageBisCheckID, &nodes, chromedp.ByID, chromedp.AtLeast(0)).Do(ctx); err != nil {
				return err
			}
			if len(nodes) == 0 {
				*result = pipres.NoRescheduleTasks
			}
			return nil
		}),
	}
}

func proceedToThirdStage(
	singleUrl *config.SingleURL,
	htmlIds *config.ElementIds,
	result *pipres.PipelineResult,
) chromedp.Tasks {
	if singleUrl.Type == "initial" {
		return chromedp.Tasks{
			chromedp.Click(htmlIds.SecondStageButtonID, chromedp.ByID),
			chromedp.ActionFunc(func(context.Context) error {
				log.WithField("name", singleUrl.Name).Println("Waiting after clicked second stage button")
				return nil
			}),
			chromedp.Sleep(1 * time.Second),
		}
	}

	return chromedp.Tasks{
		chromedp.Click(htmlIds.SecondStageBisCheckID, chromedp.ByID),
		chromedp.Click(htmlIds.SecondStageBisButtonID, chromedp.ByID),
		chromedp.ActionFunc(func(context.Context) error {
			log.WithField("name", singleUrl.Name).Println("Waiting after clicked second stage bis button")
			return nil
		}),
		chromedp.Sleep(1 * time.Second),
	}
}

func checkThirdStageResult(
	singleUrl *config.SingleURL,
	htmlIds *config.ElementIds,
	result *pipres.PipelineResult,
) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.WaitVisible("footer", chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var hasExcuseMe = regexp.MustCompile("Извините")

			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			htmlStr, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			if err != nil {
				return err
			}

			if hasExcuseMe.MatchString(htmlStr) {
				*result = pipres.SlotNotAvailable
			} else {
				*result = pipres.SlotAvailable
			}
			return nil
		}),
	}
}

// RunWholePipeline returns the result and number of retries
func RunWholePipeline(
	singleUrl *config.SingleURL,
	htmlIds *config.ElementIds,
	ctx context.Context,
) pipres.PipelineResult {
	log := log.WithField("name", singleUrl.Name)

	log.Print("Opening child chromedp context")
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	log.Print("...done opening child chromedp context")

	var (
		bypassedCaptcha bool                  = false
		result          pipres.PipelineResult = pipres.Undefined
		numRetries      int                   = 0
	)
	for !bypassedCaptcha {
		if numRetries++; numRetries > 1 {
			log.Println("CAPTCHA probably wrong, retrying..")
		}
		var captcha []byte = []byte("")
		for len(captcha) < 6 {
			var buf []byte
			if err := chromedp.Run(ctx, saveCaptchaImg(singleUrl, htmlIds, &buf)); err != nil {
				log.Fatal(err)
			}

			filename := fmt.Sprintf("imgs/captcha%d.png", singleUrl.GetHash())
			if err := ioutil.WriteFile(filename, buf, 0o644); err != nil {
				log.Fatal(err)
			}

			cap, err := exec.Command("python", "ocr.py", filename).Output()
			if err != nil {
				log.Fatal(err)
			}
			captcha = bytes.Trim(cap, "\n\t")
			if len(captcha) < 6 {
				log.Println("Decoded CAPTCHA too short, must retry")
			}
		}
		log.WithField("captcha", string(captcha)).Print("Decoded CAPTCHA")

		if err := chromedp.Run(
			ctx, submitDecodedCaptcha(singleUrl, htmlIds, &captcha, &bypassedCaptcha),
		); err != nil {
			log.Fatal(err)
		}

	}
	log.Print("Seems CAPTCHA bypassed, proceeding to the second page")
	if err := chromedp.Run(ctx, proceedToSecondStage(singleUrl, htmlIds, &result)); err != nil {
		log.Fatal(err)
	}

	if result == pipres.Undefined {
		if err := chromedp.Run(ctx, proceedToThirdStage(singleUrl, htmlIds, &result)); err != nil {
			log.Fatal(err)
		}

		if err := chromedp.Run(ctx, checkThirdStageResult(singleUrl, htmlIds, &result)); err != nil {
			log.Fatal(err)
		}
	}

	log.WithField("availability", result).WithField("requests", numRetries).Printf("Slot availability deduced")
	chromedp.Cancel(ctx)
	return result
}
