import qrcode
from config import QR_PX

def make_qr_png(content: str, out_path):
    qr = qrcode.QRCode(error_correction=qrcode.constants.ERROR_CORRECT_Q)
    qr.add_data(content)
    qr.make(fit=True)
    img = qr.make_image(fill_color="black", back_color="white").convert("RGB")
    img = img.resize((QR_PX, QR_PX))
    img.save(out_path)
    return out_path
