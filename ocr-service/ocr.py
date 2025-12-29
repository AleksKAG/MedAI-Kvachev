from flask import Flask, request, jsonify
import pytesseract
from pdf2image import convert_from_bytes
from PIL import Image
import io

app = Flask(__name__)

@app.route('/ocr', methods=['POST'])
def ocr():
    file = request.files['file']
    file_bytes = file.read()

    if file.filename.endswith('.pdf'):
        images = convert_from_bytes(file_bytes)
        text = ""
        for img in images:
            text += pytesseract.image_to_string(img, lang='rus+eng') + "\n"
    else:
        img = Image.open(io.BytesIO(file_bytes))
        text = pytesseract.image_to_string(img, lang='rus+eng')

    return jsonify({"text": text})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
