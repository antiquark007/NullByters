def generate_pdf(cert_data, output_path):
    from fpdf import FPDF

    pdf = FPDF()
    pdf.add_page()

    pdf.set_font("Arial", size=12)

    for key, value in cert_data.items():
        pdf.cell(200, 10, f"{key}: {value}", ln=True)

    pdf.output(output_path)

def generate_certificate_pdf(cert_info, output_dir):
    output_path = f"{output_dir}/certificate_{cert_info['uuid']}.pdf"
    generate_pdf(cert_info, output_path)
    return output_path

def log_pdf_generation(cert_info, log_file):
    with open(log_file, 'a') as log:
        log.write(f"Generated PDF for certificate {cert_info['uuid']} at {cert_info['timestamp']}\n")