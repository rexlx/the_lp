import pikepdf
import sys
import os
from add_image_to_pdf import add_external_image_tracker

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
        # Create a JavaScript action dictionary
        js_action = pikepdf.Dictionary(
            S=pikepdf.Name('/JavaScript'),
            JS=pikepdf.String(f"app.launchURL('{url}', true);")  # Use pikepdf.String
        )

        # Set the OpenAction
        pdf.Root.OpenAction = js_action  # Explicitly set it in the Root catalog

        # Save with debugging output
        pdf.save(output_path, linearize=True)  # Linearize for easier inspection
        print(f"Added OpenAction: {js_action}")

def add_open_xhr_action(pdf_path, output_path, url):
    with pikepdf.open(pdf_path) as pdf:
        js_action = pikepdf.Dictionary(
            S=pikepdf.Name('/JavaScript'),
            JS=pikepdf.String(f"""
                var xhr = new XMLHttpRequest();
                xhr.open('GET', '{url}', false); // Synchronous request (not recommended, but for example)
                xhr.send();
            """)
        )
        pdf.Root.OpenAction = js_action
        pdf.save(output_path, linearize=True)
        print(f"Added OpenAction: {js_action}")

def add_fetch_url_action(pdf_path, output_path, url):
    with pikepdf.open(pdf_path) as pdf:
        # JavaScript to silently fetch the URL using Net.HTTP
        js_code = f"""
        try {{
            var url = "{url}";
            Net.HTTP.request({{
                cURL: url,
                cMethod: "GET",
                bSilent: true
            }});
        }} catch (e) {{}}
        """
        js_action = pikepdf.Dictionary(
            S=pikepdf.Name('/JavaScript'),
            JS=pikepdf.String(js_code)
        )

        # Set the OpenAction
        pdf.Root.OpenAction = js_action
        pdf.save(output_path, linearize=True)
        print(f"Added OpenAction: {js_action}")

def add_image_tracker(pdf_path, output_path, url):
    with pikepdf.open(pdf_path) as pdf:
        page = pdf.pages[0]
        annot = pikepdf.Dictionary(
            Subtype=pikepdf.Name('/Widget'),
            Rect=[0, 0, 1, 1],  # 1x1 pixel, off-screen if needed
            Contents=pikepdf.String("Tracker"),
            AP=pikepdf.Dictionary(
                N=pikepdf.Dictionary(
                    Type=pikepdf.Name('/XObject'),
                    Subtype=pikepdf.Name('/Image'),
                    Width=1,
                    Height=1,
                    ColorSpace=pikepdf.Name('/DeviceRGB'),
                    BitsPerComponent=8,
                    Stream=pikepdf.Stream(pdf, b"\x00\x00\x00")  # 1x1 black pixel
                )
            ),
            A=pikepdf.Dictionary(
                S=pikepdf.Name('/URI'),
                URI=pikepdf.String(url + ".png")  # Append .png to fake an image
            )
        )
        if "/Annots" in page:
            page["/Annots"].append(annot)
        else:
            page["/Annots"] = pikepdf.Array([annot])
        pdf.save(output_path)


def add_submit_empty_form_action(pdf_path, output_path, submit_url):
    """Adds an OpenAction to submit an empty form to a URL on PDF open."""
    with pikepdf.open(pdf_path) as pdf:
        js_code = f"""
        this.submitForm("{submit_url}");
        """

        # Create a JavaScript action dictionary
        js_action = pikepdf.Dictionary(
            S=pikepdf.Name('/JavaScript'),
            JS=pikepdf.String(js_code)
        )

        # Set the OpenAction
        pdf.Root.OpenAction = js_action

        # Save the modified PDF
        pdf.save(output_path, linearize=True)
        print(f"Added OpenAction to submit form: {js_action}")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python script.py <input_pdf_path> <uuid>")
        sys.exit(1)

    pdf_path = sys.argv[1]
    uuid = sys.argv[2] # get the second argument.
    output_path = os.path.splitext(pdf_path)[0] + "_new.pdf"
    page_number = 0  # Page index (0-based)
    x, y, width, height = 100, 700, 200, 50  # Coordinates and size of the clickable area
    url = "http://fairlady:8081/" + uuid # append the uuid to the url

    add_submit_empty_form_action(pdf_path, output_path, url)
    print(f"Modified PDF saved to: {output_path}")