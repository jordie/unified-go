# Chrome Extension Icons Guide

The EduApps Chrome extension requires three PNG icon files in the `icons/` directory. These icons are displayed in Chrome's extension management page and as the extension's toolbar button.

## Required Icons

| File | Size | Purpose |
|------|------|---------|
| `icon-16.png` | 16×16 pixels | Extension page list icon |
| `icon-48.png` | 48×48 pixels | Extension page detailed view |
| `icon-128.png` | 128×128 pixels | Chrome Web Store icon |

All icons should be **PNG format** with **transparency (alpha channel)**.

## Quick Generation Methods

### Option 1: Using ImageMagick (Recommended)

If you have ImageMagick installed, create a base icon and generate all sizes:

```bash
# Create icons from a larger source image
convert source-icon.png -resize 16x16 icons/icon-16.png
convert source-icon.png -resize 48x48 icons/icon-48.png
convert source-icon.png -resize 128x128 icons/icon-128.png
```

### Option 2: Using Python PIL/Pillow

```python
from PIL import Image

# Open your icon file (should be at least 128x128)
icon = Image.open('source-icon.png')

# Create resized versions
icon.resize((16, 16), Image.LANCZOS).save('icons/icon-16.png')
icon.resize((48, 48), Image.LANCZOS).save('icons/icon-48.png')
icon.resize((128, 128), Image.LANCZOS).save('icons/icon-128.png')
```

Run:
```bash
python3 generate_icons.py
```

### Option 3: Using Online Tools

Use online icon generators:
- [Favicon Generator](https://www.favicon-generator.org/)
- [Icon Convert](https://icoconvert.com/)
- [Resize Image](https://www.resizeimage.net/)

## Design Guidelines

### Icon Style
- **Modern, flat design** with rounded corners
- **Primary colors**: Purple (#667eea) and violet (#764ba2)
- **Accent colors**: Blue (#2196F3), teal (#4ecdc4), orange (#ff6b6b)

### Icon Elements
Good icon candidates:
- Book with pen (reading + writing)
- Graduation cap (education)
- Lightbulb (learning)
- Brain (intelligence)
- Stacked books (knowledge)
- Checkmark (achievements)

### Do's and Don'ts
✅ **Do:**
- Keep it simple and recognizable
- Ensure it looks good at small sizes (16×16)
- Use transparent PNG
- Add padding/margins around the icon
- Use consistent colors

❌ **Don't:**
- Use text that's unreadable at 16×16
- Use too many colors
- Use low-resolution source images
- Forget transparency

## Example: Create an Icon from Scratch

### Using Inkscape (Free Vector Editor):

1. Create a new document (200×200px)
2. Draw your icon design
3. Export as PNG at 128×128
4. Use ImageMagick to resize to other sizes

### Using GIMP (Free Raster Editor):

1. Create new image (128×128 pixels)
2. Add your design
3. Flatten/merge layers if needed
4. File → Export As → `icon-128.png`
5. Create copies and scale down for 16×16 and 48×48

## Recommended Icon Design

Here's a simple design concept:

```
┌─────────────────────┐
│                     │
│     📚 + ✏️ + 🧠    │
│                     │
│  (Books + Writing   │
│   + Intelligence)   │
│                     │
└─────────────────────┘
```

Colors:
- Background: Transparent
- Main elements: #667eea (primary) and #764ba2 (secondary)
- Accents: #4ecdc4 (teal)

## Testing Icons

After creating your icons, test them:

1. Load the extension in developer mode
2. Check Chrome's extension page to see the icon
3. Verify all three sizes render correctly
4. Test on different desktop backgrounds

If icons look blurry or pixelated:
- Use a larger source image
- Try different resize algorithms (LANCZOS, BICUBIC)
- Verify all icons are PNG with transparency

## Placeholder Icons

For development, you can use simple placeholder icons:

```bash
# Create solid-color placeholder PNGs
python3 << 'EOF'
from PIL import Image, ImageDraw

def create_placeholder(size, filename):
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)

    # Draw gradient-like effect
    for i in range(size):
        r = int(102 * (i / size))
        g = int(126 * (i / size))
        b = int(234 * (i / size))
        draw.rectangle([(0, i), (size, i+1)], fill=(r, g, b, 255))

    img.save(filename)

create_placeholder(16, 'icons/icon-16.png')
create_placeholder(48, 'icons/icon-48.png')
create_placeholder(128, 'icons/icon-128.png')

print('Placeholder icons created')
EOF
```

## Submitting to Chrome Web Store

When publishing to the Chrome Web Store, ensure:

1. ✅ All three icon files exist and are correct size
2. ✅ Icons are PNG format with transparency
3. ✅ Icon is professional and represents the extension
4. ✅ Small icons (16×16) remain recognizable
5. ✅ Compliance with Chrome Web Store design guidelines

## Troubleshooting

**Icons not appearing:**
- Verify files are in `icons/` directory
- Check manifest.json references (icon-16.png, not icon_16.png)
- Reload extension in chrome://extensions/

**Icons look blurry:**
- Increase source image resolution
- Use LANCZOS or BICUBIC scaling
- Create 256×256 source and scale down

**Icons too bright/dark:**
- Adjust color values in your source image
- Ensure contrast ratio is at least 4.5:1

## Resources

- [Chrome Extension Icon Guidelines](https://developer.chrome.com/docs/extensions/reference/manifest/icons/)
- [PNG Optimization Tools](https://tinypng.com/)
- [Color Accessibility Checker](https://webaim.org/resources/contrastchecker/)
- [Icon Design Inspiration](https://www.flaticon.com/)

---

**Note:** Before distribution on the Chrome Web Store, ensure your icons comply with their guidelines and policies.
