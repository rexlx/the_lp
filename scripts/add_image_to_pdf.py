import pikepdf
import uuid # For generating a unique resource name

def add_external_image_tracker(pdf_path, output_path, url):
    try:
        with pikepdf.open(pdf_path) as pdf:
            if not pdf.pages:
                print(f"Error: PDF '{pdf_path}' contains no pages.")
                return False

            # --- Target the first page ---
            page = pdf.pages[0]

            # --- 1. Create the File Specification Dictionary pointing to the URL ---
            # This dictionary tells the PDF viewer where to find the external resource.
            filespec = pikepdf.Dictionary(
                Type=pikepdf.Name.Filespec,
                # F=pikepdf.String("tracker.gif"), # Optional: dummy filename
                URL=pikepdf.String(url)         # The actual tracking URL
            )
            # It's good practice to make complex objects indirect
            filespec_ref = pdf.make_indirect(filespec)

            # --- 2. Create the Image XObject Dictionary ---
            # This defines an 'image' but points to the external filespec for data.
            # We use minimal valid image properties (1x1 grayscale).
            # Generate a unique name to avoid conflicts in page resources
            tracker_img_name = pikepdf.Name(f"/TrackerImg_{uuid.uuid4().hex[:8]}")

            image_xobject = pikepdf.Dictionary(
                Type=pikepdf.Name.XObject,
                Subtype=pikepdf.Name.Image,
                Width=1, # Minimal dimensions
                Height=1,
                ColorSpace=pikepdf.Name.DeviceGray, # Minimal colorspace
                BitsPerComponent=1,
                # Crucially, link to the external file specification:
                F=filespec_ref
                # No Filter or image data Stream needed here; data is external via /F -> /URL
            )
            image_xobject_ref = pdf.make_indirect(image_xobject)

            # --- 3. Add the Image XObject to the Page's Resources ---
            # Ensure Resources and XObject dictionaries exist
            if not hasattr(page, 'Resources') or page.Resources is None:
                page.Resources = pikepdf.Dictionary()
            if not hasattr(page.Resources, 'XObject') or page.Resources.XObject is None:
                page.Resources.XObject = pikepdf.Dictionary()

            # Add our new image resource definition
            page.Resources.XObject[tracker_img_name] = image_xobject_ref
            print(f"Added Image XObject '{tracker_img_name}' to Page 1 Resources.")

            # --- 4. Modify the Page's Content Stream to use the image ---
            # This 'Do' command instructs the renderer to draw the image,
            # which *might* trigger the external URL load attempt.
            # We save/restore graphics state (q/Q) and place the 1x1 image
            # at coordinates (-10, -10) relative to default origin, likely off-page.
            img_name_bytes = tracker_img_name.name.encode('latin-1')
            # Command Breakdown: q=save state, cm=set matrix (move origin), Do=draw, Q=restore state
            new_command = b"\nq\n1 0 0 1 -10 -10 cm\n" + img_name_bytes + b" Do\nQ\n"

            # Append the command safely, handling existing single/multiple content streams
            content_stream_data = b""
            if hasattr(page, 'Contents') and page.Contents is not None:
                 if isinstance(page.Contents, pikepdf.Array):
                     # Concatenate existing streams if page uses an array
                     for stream in page.Contents:
                         content_stream_data += stream.read_bytes() + b"\n"
                 elif isinstance(page.Contents, pikepdf.Stream):
                     content_stream_data = page.Contents.read_bytes()
                 else:
                    print("Warning: Page Contents is not an Array or Stream. Treating as empty.")
                    content_stream_data = b""
                 # Else: object exists but is not Stream or Array (unlikely, treat as empty)

            # Create a new stream with original data + new command
            new_stream_data = content_stream_data + new_command
            new_content_stream = pikepdf.Stream(pdf, new_stream_data)
            # Replace old Contents with the new stream (making it indirect)
            page.Contents = pdf.make_indirect(new_content_stream)

            print(f"Appended command to draw '{tracker_img_name}' to Page 1 Contents.")

            # Optional: Remove any existing OpenAction if you want this to be the only mechanism
            if hasattr(pdf.Root, 'OpenAction'):
                 try:
                     del pdf.Root.OpenAction
                     print("Removed existing /OpenAction from Root.")
                 except KeyError:
                     pass # Ignore if it wasn't actually there

            # --- 5. Save the modified PDF ---
            # Linearize=True can sometimes help with compatibility but isn't strictly necessary
            pdf.save(output_path, linearize=False)
            print(f"Successfully attempted to add external image tracker. Modified PDF saved to '{output_path}'")
            return True

    except pikepdf.PasswordError:
        print(f"Error: PDF '{pdf_path}' is password-protected.")
        return False
    except Exception as e:
        print(f"An unexpected error occurred: {e}")
        import traceback
        traceback.print_exc()
        return False