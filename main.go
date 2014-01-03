package main

import (
	"errors"
	"fmt"
	"github.com/Ferguzz/glam"
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

const vertex_shader = `
	#version 150

	in vec2 position;
	in vec2 textureCoord;
	out vec2 TextureCoord;
	uniform mat4 model;
	uniform mat4 view;
	uniform mat4 projection;

	void main()
	{
	    TextureCoord = textureCoord;
	    gl_Position = projection * view * model * vec4(position, 0.0, 1.0);
	}
	`

const fragment_shader = `
	#version 150

	out vec4 outColor;
	in vec2 TextureCoord;
	uniform sampler2D tex1;
	uniform sampler2D tex2;

	void main()
	{
	    outColor = texture(tex1, TextureCoord);
	}
`

var rotate bool = true

func errorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("%v: %v\n", err, desc)
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeyEscape, glfw.KeyQ:
			window.SetShouldClose(true)
		case glfw.KeyR:
			rotate = !rotate
		}
	}
}

func loadShader(shaderType gl.GLenum, source string) (gl.Shader, error) {
	shader := gl.CreateShader(shaderType)
	shader.Source(source)
	shader.Compile()

	if shader.Get(gl.COMPILE_STATUS) == 0 {
		return shader, errors.New(fmt.Sprintf("Shader (%v) did not compile.", source))
	}

	return shader, nil
}

func loadImage(filename string) (*image.NRGBA, error) {
	src, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	nrgbaImg, ok := img.(*image.NRGBA)
	if !ok {
		nrgbaImg = image.NewNRGBA(image.Rect(0, 0, width, height))
		draw.Draw(nrgbaImg, nrgbaImg.Bounds(), img, bounds.Min, draw.Src)
	}

	return nrgbaImg, nil
}

func loadTextureFromImage(filename string, texture_id gl.GLenum) (gl.Texture, error) {
	texture := gl.GenTexture()
	image, err := loadImage(filename)
	if err != nil {
		return texture, err
	}
	dims := image.Bounds()

	gl.ActiveTexture(texture_id)
	texture.Bind(gl.TEXTURE_2D)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, dims.Dx(), dims.Dy(), 0, gl.RGBA, gl.UNSIGNED_BYTE, image.Pix)

	return texture, nil
}

func glInit() (*glfw.Window, error) {
	glfw.SetErrorCallback(errorCallback)

	if !glfw.Init() {
		return nil, errors.New("Can't initialise GLFW!")
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile)
	glfw.WindowHint(glfw.OpenglForwardCompatible, gl.TRUE)

	window, err := glfw.CreateWindow(640, 480, "OpenGL Textures", nil, nil)
	if err != nil {
		return nil, err
	}

	window.SetKeyCallback(keyCallback)
	window.MakeContextCurrent()
	if gl.Init() != 0 {
		return nil, errors.New("Can't initialise OpenGL.")
	}

	return window, nil
}

func glExit() {
	glfw.Terminate()
}

func main() {
	window, err := glInit()
	if err != nil {
		panic(err)
	}
	defer glExit()

	vertices := []gl.GLfloat{-0.5, -0.5, 0, 1, 0.5, -0.5, 1, 1, -0.5, 0.5, 0, 0, 0.5, 0.5, 1, 0}
	elements := []gl.GLushort{0, 1, 2, 3}

	vao := gl.GenVertexArray()
	vao.Bind()
	vbo := gl.GenBuffer()
	vbo.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, vertices, gl.STATIC_DRAW)
	elementBuffer := gl.GenBuffer()
	elementBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(elements)*2, elements, gl.STATIC_DRAW)

	defer vao.Delete()
	defer vbo.Delete()
	defer elementBuffer.Delete()

	texture1, err := loadTextureFromImage("sloth_n_banana.jpg", gl.TEXTURE0)
	if err != nil {
		panic(err)
	}
	defer texture1.Delete()
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	texture2, err := loadTextureFromImage("sloth_n_kebab.jpg", gl.TEXTURE1)
	if err != nil {
		panic(err)
	}
	defer texture2.Delete()
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	vertexShader, err := loadShader(gl.VERTEX_SHADER, vertex_shader)
	if err != nil {
		panic(err)
	}
	defer vertexShader.Delete()

	fragmentShader, err := loadShader(gl.FRAGMENT_SHADER, fragment_shader)
	if err != nil {
		panic(err)
	}
	defer fragmentShader.Delete()

	shaderProgram := gl.CreateProgram()
	shaderProgram.AttachShader(vertexShader)
	shaderProgram.AttachShader(fragmentShader)
	shaderProgram.BindFragDataLocation(0, "outColor")
	shaderProgram.Link()
	shaderProgram.Use()
	defer shaderProgram.Delete()

	positionAttrib := shaderProgram.GetAttribLocation("position")
	positionAttrib.AttribPointer(2, gl.FLOAT, false, 4*4, uintptr(0))
	positionAttrib.EnableArray()

	textureCorrdAttrib := shaderProgram.GetAttribLocation("textureCoord")
	textureCorrdAttrib.AttribPointer(2, gl.FLOAT, false, 4*4, (uintptr(2 * 4)))
	textureCorrdAttrib.EnableArray()

	shaderProgram.GetUniformLocation("tex1").Uniform1i(0)
	shaderProgram.GetUniformLocation("tex2").Uniform1i(1)

	// startTime := time.Now()

	modelMat := glam.Identity()
	projectionMat := glam.Perspective(45, 640/480, 1, 10)

	shaderProgram.GetUniformLocation("projection").UniformMatrix4fv(false, projectionMat)
	spinCount := 0.0

	for !window.ShouldClose() {
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		// clock := time.Since(startTime).Seconds()

		if rotate {
			modelMat = glam.Rotation(float32(spinCount*math.Pi), glam.Vec3{0, 0, 1})
			spinCount += 0.0005
		}
		shaderProgram.GetUniformLocation("model").UniformMatrix4fv(false, modelMat)

		viewMat := glam.LookAt(glam.Vec3{1, 0, 1}, glam.Vec3{0, 0, 0}, glam.Vec3{0, 0, 1})
		shaderProgram.GetUniformLocation("view").UniformMatrix4fv(false, viewMat)

		gl.DrawElements(gl.TRIANGLE_STRIP, 4, gl.UNSIGNED_SHORT, uintptr(0))

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
