from reportlab.lib.pagesizes import A4
from reportlab.lib import colors
from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer, Image
from reportlab.lib.styles import getSampleStyleSheet
from pathlib import Path
from PyPDF2 import PdfReader, PdfWriter
from PyPDF2.generic import NameObject, createStringObject
import json

styles = getSampleStyleSheet()

def make_section(title, data, col_widths=None):
    section_title = Paragraph(f"<b>{title}</b>", styles['Heading4'])
    table = Table(data, colWidths=col_widths)
    table.setStyle(TableStyle([
        ('BOX', (0,0), (-1,-1), 0.5, colors.black),
        ('INNERGRID', (0,0), (-1,-1), 0.25, colors.grey),
        ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
        ('FONTNAME', (0,0), (-1,-1), 'Helvetica'),
        ('FONTSIZE', (0,0), (-1,-1), 9),
        ('LEFTPADDING', (0,0), (-1,-1), 4),
        ('RIGHTPADDING', (0,0), (-1,-1), 4),
    ]))
    return [section_title, table, Spacer(1, 6)]

def header(elements, subtitle: str | None = None):
    elements.append(Paragraph("<b><font size=16>CERTIFICATE OF SANITIZATION</font></b>", styles['Title']))
    elements.append(Spacer(1, 6))
    if subtitle:
        elements.append(Paragraph(f"<font size=10>{subtitle}</font>", styles['Normal']))
        elements.append(Spacer(1, 12))
    else:
        elements.append(Spacer(1, 12))

def generate_certificate_pdf(cert_data: dict,
                              qr_png_path: Path,
                              qr_url: str,
                              output_pdf_path: Path,
                              subtitle: str | None = "Issued by NullBytes",
                              payload_obj: dict | None = None):
    # --- Build PDF with ReportLab ---
    doc = SimpleDocTemplate(str(output_pdf_path), pagesize=A4,
                            rightMargin=20, leftMargin=20, topMargin=20, bottomMargin=20)
    elements = []
    header(elements, subtitle)

    # Sections
    person = cert_data["performer"]
    elements.extend(make_section("Person Performing Sanitization", [
        ["Name:", person["name"], "Title:", person["title"]],
        ["Organization:", person["organization"], "Location:", person["location"]],
        ["Phone:", person["phone"], "", ""],
    ], [80, 150, 80, 150]))

    media = cert_data["media"]
    elements.extend(make_section("Media Information", [
        ["Make/Vendor:", media["vendor"], "Model:", media["model"]],
        ["Serial Number:", media["serial_number"], "Property Number:", media["property_number"]],
        ["Media Type:", media["type"], "Source:", media["source"]],
        ["Classification:", media["classification"], "Data Backed Up:", media["data_backed_up"]],
    ], [100, 130, 100, 130]))

    s = cert_data["sanitization"]
    elements.extend(make_section("Sanitization Details", [
        ["Method Type:", s["method_type"], "Method Used:", s["method_used"]],
        ["Tool Used (version):", s["tool_used"], "Verification Method:", s["verification_method"]],
        ["Post-Sanitization Classification:", s["post_classification"], "", ""],
    ], [120, 110, 120, 110]))

    d = cert_data["destination"]
    elements.extend(make_section("Media Destination", [
        ["Destination:", d["type"], "Details:", d["details"]],
    ], [100, 130, 100, 130]))

    # QR block
    qr_img = Image(str(qr_png_path), width=110, height=110)
    qr_table = Table([
        [qr_img],
        [Paragraph("<b>Scan here for verification</b>", styles['BodyText'])],
        [Paragraph("Opens verifier site and auto-verifies", styles['BodyText'])]
    ], colWidths=[140])
    qr_table.setStyle(TableStyle([
        ('BOX', (0,0), (-1,-1), 0.5, colors.black),
        ('ALIGN', (0,0), (-1,-1), 'CENTER'),
        ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
        ('TOPPADDING', (0,0), (-1,-1), 6),
        ('BOTTOMPADDING', (0,0), (-1,-1), 6),
    ]))
    elements.append(Spacer(1, 12))
    elements.append(qr_table)
    elements.append(Spacer(1, 4))
    elements.append(Paragraph(f"<font size=6 color='white'>{qr_url}</font>", styles['Normal']))

    doc.build(elements)

    # --- Embed payload into PDF metadata ---
    if payload_obj:
        reader = PdfReader(str(output_pdf_path))
        writer = PdfWriter()
        writer.append_pages_from_reader(reader)

        payload_json = json.dumps(payload_obj, separators=(",", ":"), sort_keys=True)
        writer.add_metadata({
            NameObject("/CertPayload"): createStringObject(payload_json)
        })

        with open(output_pdf_path, "wb") as f_out:
            writer.write(f_out)
