import pikepdf
import sys
import os

def add_url_action(pdf_path, output_path, page_number, x, y, width, height, url):
    with pikepdf.open(pdf_path) as pdf:
        page = pdf.pages[page_number]

        # Create the annotation dictionary directly.
        annot = pikepdf.Dictionary(
            Subtype=pikepdf.Name('/Link'),
            Rect=[x, y, x + width, y + height],
            A=pikepdf.Dictionary(S=pikepdf.Name('/URI'), URI=url)
        )

        # Append the dictionary to the page's annotations.
        if "/Annots" in page:
            page["/Annots"].append(annot)
        else:
            page["/Annots"] = pikepdf.Array([annot])

        pdf.save(output_path)

def add_open_url_action(pdf_path, output_path, url):
    with pikepdf.open(pdf_path) as pdf:
        # Create a JavaScript action dictionary.
        js_action = pikepdf.Dictionary(
            S=pikepdf.Name('/JavaScript'),
            JS=f"app.launchURL('{url}', true);"  # 'true' opens in a new window/tab
        )

        # Set the OpenAction for the PDF document.
        pdf.OpenAction = js_action

        pdf.save(output_path)

if __name__ == "__main__":
    if len(sys.argv) != 3:  # Expecting 3 arguments now
        print("Usage: python script.py <input_pdf_path> <uuid>")
        sys.exit(1)

    pdf_path = sys.argv[1]
    uuid = sys.argv[2] # get the second argument.
    output_path = os.path.splitext(pdf_path)[0] + "_new.pdf"
    page_number = 0  # Page index (0-based)
    x, y, width, height = 100, 700, 200, 50  # Coordinates and size of the clickable area
    url = "http://fairlady:8081/" + uuid # append the uuid to the url

    add_open_url_action(pdf_path, output_path, page_number, x, y, width, height, url)
    print(f"Modified PDF saved to: {output_path}")