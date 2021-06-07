import re
import sys
import os
import binascii

from functools import partial

import cv2
import numpy as np
import pytesseract


# get grayscale image
def get_grayscale(image):
    return cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)


# noise removal
def remove_noise(image):
    return cv2.medianBlur(image, 3)


# thresholding
def thresholding(image):
    return cv2.threshold(image, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)[1]


# dilation
def dilate(image):
    kernel = np.ones((5, 5), np.uint8)
    return cv2.dilate(image, kernel, iterations=1)


# erosion
def erode(image):
    kernel = np.ones((5, 5), np.uint8)
    return cv2.erode(image, kernel, iterations=1)


# closing - dilation followed by erosion
def closing(image):
    kernel = np.ones((3, 3), np.uint8)
    return cv2.morphologyEx(image, cv2.MORPH_CLOSE, kernel)


# opening - erosion followed by dilation
def opening(image):
    kernel = np.ones((3, 3), np.uint8)
    return cv2.morphologyEx(image, cv2.MORPH_OPEN, kernel)


# canny edge detection
def canny(image):
    return cv2.Canny(image, 100, 200)


# skew correction
def deskew(image):
    coords = np.column_stack(np.where(image > 0))
    angle = cv2.minAreaRect(coords)[-1]
    if angle < -45:
        angle = -(90 + angle)
    else:
        angle = -angle
    (h, w) = image.shape[:2]
    center = (w // 2, h // 2)
    M = cv2.getRotationMatrix2D(center, angle, 1.0)
    rotated = cv2.warpAffine(image, M, (w, h), flags=cv2.INTER_CUBIC, borderMode=cv2.BORDER_REPLICATE)
    return rotated


# template matching
def match_template(image, template):
    return cv2.matchTemplate(image, template, cv2.TM_CCOEFF_NORMED)


def transform0(img):
    img = get_grayscale(img)
    img = remove_noise(img)
    img = thresholding(img)
    img = opening(img)
    img = closing(img)

    return img


def transform1(img, shear_scale):
    img = get_grayscale(img)
    img = remove_noise(img)
    img = thresholding(img)
    img = opening(img)
    img = closing(img)

    height, width = img.shape[:2]
    dx = int(shear_scale * width)

    box0 = np.array(
        [
            [0, 0],
            [width, 0],
            [width, height],
            [0, height],
        ],
        np.float32,
    )
    box1 = np.array(
        [
            [+dx, 0],
            [width + dx, 0],
            [width - dx, height],
            [-dx, height],
        ],
        np.float32,
    )

    box0 = box0.astype(np.float32)
    box1 = box1.astype(np.float32)
    mat = cv2.getPerspectiveTransform(box0, box1)

    img = cv2.warpPerspective(
        img,
        mat,
        (width, height),
        flags=cv2.INTER_LINEAR,
        borderMode=cv2.BORDER_CONSTANT,
        borderValue=(255),
    )

    return img


def main():
    for t in [
        transform0,
        partial(transform1, shear_scale=0.05),
        partial(transform1, shear_scale=-0.5),
        partial(transform1, shear_scale=0.075),
        partial(transform1, shear_scale=-0.075),
        partial(transform1, shear_scale=0.1),
        partial(transform1, shear_scale=-0.1),
    ]:
        img = cv2.imread(sys.argv[1])
        img = cv2.resize(img, (280, 70), interpolation=cv2.INTER_CUBIC)
        img = t(img)

        custom_config = r"--psm 8 --oem 3 -c tessedit_char_whitelist=0123456789"
        res = pytesseract.image_to_string(img, config=custom_config)
        match = re.match(r"\d{6,6}", res)
        if match:
            print(match.group())
            return
        else:
            # Uncomment to enable saving wrong images (e.g. to later inspect them)
            # cv2.imwrite(sys.argv[1][:-4] + ".wrongmatch-" + str(binascii.hexlify(os.urandom(20))) + ".png", img)
            continue


if __name__ == "__main__":
    main()
