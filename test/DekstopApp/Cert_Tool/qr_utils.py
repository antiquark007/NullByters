def generate_qr_code(data):
    import qrcode
    qr = qrcode.QRCode(
        version=1,
        error_correction=qrcode.constants.ERROR_CORRECT_L,
        box_size=10,
        border=4,
    )
    qr.add_data(data)
    qr.make(fit=True)
    img = qr.make_image(fill_color="black", back_color="white")
    return img

def save_qr_code(img, file_path):
    img.save(file_path)

def read_qr_code(file_path):
    from PIL import Image
    from pyzbar.pyzbar import decode
    img = Image.open(file_path)
    decoded_objects = decode(img)
    return [obj.data.decode('utf-8') for obj in decoded_objects]

def generate_qr_for_certificate(cert_data, output_path):
    img = generate_qr_code(cert_data)
    save_qr_code(img, output_path)

def verify_qr_code(file_path):
    return read_qr_code(file_path)