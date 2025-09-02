from PIL import Image, ImageDraw, ImageFont

# Load and resize the image
image = Image.open("buddha.jpg").convert("RGB")
char_width = 100
aspect_ratio = image.height / image.width
char_height = int(char_width * aspect_ratio)
image = image.resize((char_width, char_height))

# Create canvas
canvas = Image.new("RGB", image.size, "white")
draw = ImageDraw.Draw(canvas)
font = ImageFont.truetype("DejaVuSansMono.ttf", size=10)

# Characters for shading
characters = "@%#*+=-:. "

for y in range(image.height):
    for x in range(image.width):
        r, g, b = image.getpixel((x, y))
        brightness = int(0.299 * r + 0.587 * g + 0.114 * b)
        char = characters[brightness * (len(characters) - 1) // 255]
        draw.text((x, y), char, fill=(r, g, b), font=font)

# Save the final image
canvas.save("buddha_ascii_colored.png")
