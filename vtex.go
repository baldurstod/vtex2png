package main

import (
	//"os"
	"io"
	//"log"
	"fmt"
	"bytes"
	"math"
	"github.com/pierrec/lz4/v4"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

type VtexHeader struct {
	FileLength int32
	FileVersion [2]int16
	ResourceOffset int32
	ResourceCount int32
}

func (this *VtexHeader) GetResourceOffset() uint32 {
	return uint32(this.ResourceOffset + 8)
}

type VtexFile struct {
	datas []byte
	Header VtexHeader
	resources []VtexResource
}

type VtexResource struct {
	offset uint32
	ResType [4]byte
	ResourceOffset int32
	ResourceLength int32
}

type VtexExtraData struct {
	offset uint32
	ExtraDataType uint32
	ExtraDataOffset uint32
	ExtraDataLength uint32
}

func (this *VtexExtraData) GetExtraDataOffset() uint32 {
	return this.ExtraDataOffset + this.offset + 4
}

type VtexDataBlock struct {
	offset uint32
	Data VtexData
}

func (this *VtexDataBlock) GetExtraDataOffset() uint32 {
	return this.Data.ExtraDataOffset + this.offset + 32
}

type VtexData struct {
	Version uint16
	Flags uint16
	Reflectivity [16]byte
	Width uint16
	Height uint16
	Depth uint16

	ImageFormat uint8
	NumMipMap uint8

	Picmip0Res uint32
	ExtraDataOffset uint32
	ExtraDataCount uint32


	A [8]byte

	Unknown4 uint32
	Unknown5 uint32

	BumpMapScale uint32

/*
	unsigned short	vtex_version;
	unsigned short	vtex_flags;
	char 			vtex_reflectivity[16];
	unsigned short	vtex_width;
	unsigned short	vtex_height;
	unsigned short	vtex_depth;

	unsigned char	vtex_image_format;
	unsigned char	vtex_num_mipLevels;
	unsigned int	vtex_unknown1;
	unsigned int	vtex_unknown2;
	unsigned int	vtex_unknown3;
	char 			a[8];
	unsigned int	vtex_unknown4;
	unsigned int	vtex_unknown5;
	unsigned int	vtex_bump_map_scale;
*/
}

func (this *VtexData) GetImageSize(mipmapWidth int, mipmapHeight int) uint32 {
	switch int(this.ImageFormat) {
		case VTEX_FORMAT_DXT1:
			return uint32(math.Max(float64(mipmapWidth * mipmapHeight) * 0.5, 8)); // 0.5 byte per pixel
		case VTEX_FORMAT_DXT5:
			return uint32(math.Max(float64(mipmapWidth), 4) * math.Max(float64(mipmapHeight), 4)); // 1 byte per pixel
		case VTEX_FORMAT_R8:
			return uint32(math.Max(float64(mipmapWidth), 1) * math.Max(float64(mipmapHeight), 1)); // 1 byte per pixel;
		case VTEX_FORMAT_R8G8B8A8_UINT, VTEX_FORMAT_BGRA8888:
			return uint32(mipmapWidth * mipmapHeight * 4);// 4 bytes per pixel
		case VTEX_FORMAT_PNG_R8G8B8A8_UINT:
			return ^uint32(0)
		case VTEX_FORMAT_BC4:
			return uint32(math.Ceil(float64(mipmapWidth) / 4) * math.Ceil(float64(mipmapHeight) / 4) * 8);// 0.5 byte per pixel
		case VTEX_FORMAT_BC7:
			return uint32(math.Max(float64(mipmapWidth), 4) * math.Max(float64(mipmapHeight), 4));// 1 byte per pixel, blocks of 16 bytes
		default:
			panic("Unknown image format")

/*
		case VTEX_FORMAT_DXT1:
			entrySize = Math.max(mipmapWidth * mipmapHeight * 0.5, 8); // 0.5 byte per pixel
			break;
		case VTEX_FORMAT_DXT5:
			entrySize = Math.max(mipmapWidth, 4) * Math.max(mipmapHeight, 4); // 1 byte per pixel
			break;
		case VTEX_FORMAT_R8:
			entrySize = Math.max(mipmapWidth, 1) * Math.max(mipmapHeight, 1); // 1 byte per pixel;
			break;
		case VTEX_FORMAT_R8G8B8A8_UINT:
		case VTEX_FORMAT_BGRA8888:
			// 4 bytes per pixel
			entrySize = mipmapWidth * mipmapHeight * 4;
			break;
		case VTEX_FORMAT_PNG_R8G8B8A8_UINT:
			entrySize = reader.byteLength - reader.tell();
			let a = reader.tell();
			//SaveFile('loadout.obj', b64toBlob(encode64(reader.getString(entrySize))));//TODOv3: removeme
			reader.seek(a);
			break;
		case VTEX_FORMAT_BC4:
			entrySize = Math.ceil(mipmapWidth / 4) * Math.ceil(mipmapHeight / 4) * 8;// 0.5 byte per pixel
			break;
		case VTEX_FORMAT_BC7:
			entrySize = Math.max(mipmapWidth, 4) * Math.max(mipmapHeight, 4);// 1 byte per pixel, blocks of 16 bytes
			break;
			*/
	}
}

func (this *VtexResource) GetResourceType() string {
	return string(this.ResType[:])
}

func (this *VtexResource) GetResourceOffset() uint32 {
	return uint32(this.ResourceOffset) + this.offset + 4
}


func (this *VtexFile) SetData(datas []byte) {
	this.datas = datas
	reader := bytes.NewReader(datas)

	err := binary.Read(reader, binary.LittleEndian, &this.Header)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}

	this.resources = make([]VtexResource, this.Header.ResourceCount)
	currentOffset := this.Header.GetResourceOffset()
	reader.Seek(int64(currentOffset), io.SeekStart)
	for i := int32(0); i < this.Header.ResourceCount; i++ {
		this.resources[i].offset = currentOffset

		binary.Read(reader, binary.LittleEndian, &this.resources[i].ResType)
		binary.Read(reader, binary.LittleEndian, &this.resources[i].ResourceOffset)
		binary.Read(reader, binary.LittleEndian, &this.resources[i].ResourceLength)



		//fmt.Println(this.resources[i].GetResourceType(), this.resources[i].GetResourceOffset())
		currentOffset += 12
	}

	//fmt.Println(this.Header, this.resources)
}


func (this *VtexFile) GetVtexData() []byte {
	reader := bytes.NewReader(this.datas)

	var compressedMips []uint32
	var compressedMipsCount uint32
	var buffer []byte

	for i := int32(0); i < this.Header.ResourceCount; i++ {
		res := &this.resources[i]
		if (res.GetResourceType() == "DATA") {
			//fmt.Println(res, res.GetResourceOffset())
			reader.Seek(int64(res.GetResourceOffset()), io.SeekStart)
			vtexDataBlock := VtexDataBlock{}
			vtexDataBlock.offset = res.GetResourceOffset()
			vtexData := &vtexDataBlock.Data
			binary.Read(reader, binary.LittleEndian, vtexData)
			//fmt.Println("VTEX data", vtexData)

			vtexExtraDatas := make([]VtexExtraData, vtexData.ExtraDataCount)
			offset := vtexDataBlock.GetExtraDataOffset()
			reader.Seek(int64(offset), io.SeekStart)

			//fmt.Println("offset", offset)
			for extraDataIndex := uint32(0); extraDataIndex < vtexData.ExtraDataCount; extraDataIndex++ {
				vtexExtraData := &vtexExtraDatas[extraDataIndex]
				vtexExtraData.offset = offset
				offset += 12
				binary.Read(reader, binary.LittleEndian, &vtexExtraData.ExtraDataType)
				binary.Read(reader, binary.LittleEndian, &vtexExtraData.ExtraDataOffset)
				binary.Read(reader, binary.LittleEndian, &vtexExtraData.ExtraDataLength)
			}

			for extraDataIndex := uint32(0); extraDataIndex < vtexData.ExtraDataCount; extraDataIndex++ {
				vtexExtraData := &vtexExtraDatas[extraDataIndex]

				if (vtexExtraData.ExtraDataType == VTEX_EXTRA_DATA_TYPE_COMPRESSED_MIP_SIZE) {
					reader.Seek(int64(vtexExtraData.GetExtraDataOffset()), io.SeekStart)
					//fmt.Println("Extra data", vtexExtraData, vtexExtraData.GetExtraDataOffset())

					var unk1 uint32
					var unk2 uint32

					binary.Read(reader, binary.LittleEndian, &unk1)
					binary.Read(reader, binary.LittleEndian, &unk2)
					binary.Read(reader, binary.LittleEndian, &compressedMipsCount)

					compressedMips = make([]uint32, compressedMipsCount)
					//fmt.Println("compressedMipsCount", compressedMipsCount)

					for mipsIndex := uint32(0); mipsIndex < compressedMipsCount; mipsIndex++ {
						binary.Read(reader, binary.LittleEndian, &compressedMips[mipsIndex])
					}
					//fmt.Println(compressedMips)
				}
			}


			//fmt.Println(vtexExtraDatas);

			faceCount := 1;
			if ((vtexData.Flags & uint16(VTEX_FLAG_CUBE_TEXTURE)) == uint16(VTEX_FLAG_CUBE_TEXTURE)) {
				faceCount = 6;
			}

			mipmapWidth := int(float64(vtexData.Width) * math.Pow(0.5, float64(vtexData.NumMipMap) - 1));
			mipmapHeight := int(float64(vtexData.Height) * math.Pow(0.5, float64(vtexData.NumMipMap) - 1));
			var compressedMipsIndex = 0


			reader.Seek(int64(this.Header.FileLength), io.SeekStart)
			for mipIndex := uint8(0); mipIndex < vtexData.NumMipMap; mipIndex++ {
			//for(int resIndex = 0; resIndex < vtex->vtex_num_mipLevels; resIndex++) {
				//Todo : add frame support + depth support
				for faceIndex := 0; faceIndex < faceCount; faceIndex++ {
					//get_image(resource_header, vtex, mipmapWidth, mipmapHeight, data, filesize, out_path);
					size := vtexData.GetImageSize(mipmapWidth, mipmapHeight)

					compressedMipSize := size
					if (compressedMipsIndex < len(compressedMips)) {
						compressedMipSize = compressedMips[compressedMipsIndex]
					}


					if (size == compressedMipSize) {
						// No compression
						if (size == ^uint32(0)) { // read until eof
							size = uint32(reader.Len())
						}

						buffer = make([]byte, size)
						n, _ := reader.Read(buffer)
						if (n != int(size)) {
							fmt.Println("cannot read bytes", size, n)
							panic("cannot read bytes")
						}
					} else {
						// lz4 compression
						compressedBuffer := make([]byte, compressedMipSize)
						buffer = make([]byte, size)
						n, _ := reader.Read(compressedBuffer)
						if (n != int(compressedMipSize)) {
							fmt.Println("cannot read compressed bytes", compressedMipSize, n)
							panic("cannot read compressed bytes")
						}

						lz4.UncompressBlock(compressedBuffer, buffer)
					}

					decodeBuffer(&buffer, vtexData.ImageFormat, mipmapWidth, mipmapHeight)

					switch int(vtexData.ImageFormat) {
						case VTEX_FORMAT_DXT1, VTEX_FORMAT_DXT5, VTEX_FORMAT_R8, VTEX_FORMAT_R8G8B8A8_UINT, VTEX_FORMAT_BGRA8888, VTEX_FORMAT_BC4, VTEX_FORMAT_BC7:
							// Encode to png

							img := image.NewNRGBA(image.Rect(0, 0, mipmapWidth, mipmapHeight))

							bufferIndex := 0
							for y := 0; y < mipmapHeight; y++ {
								for x := 0; x < mipmapWidth; x++ {
									img.Set(x, y, color.NRGBA{
										R: uint8(buffer[bufferIndex + 0]),
										G: uint8(buffer[bufferIndex + 1]),
										B: uint8(buffer[bufferIndex + 2]),
										A: uint8(buffer[bufferIndex + 3]),
									})
									bufferIndex += 4
								}
							}

							pngBuffer := bytes.Buffer{}
							png.Encode(&pngBuffer, img)
							buffer = pngBuffer.Bytes()

						case VTEX_FORMAT_PNG_R8G8B8A8_UINT:
							// Nothing to do
						default:
							panic("Unknown image format")
					}

					//fmt.Println(size, compressedMipSize, len(compressedMips))

					compressedMipsIndex++
				}
				mipmapWidth *= 2;
				mipmapHeight *= 2;
			}

			//fmt.Println(vtexData, faceCount, mipmapWidth, mipmapHeight);
		}
	}
	return buffer
}

func decodeBuffer(buffer *[]byte, imageFormat uint8, width int, height int) {
	switch int(imageFormat) {
		case VTEX_FORMAT_DXT1, VTEX_FORMAT_DXT5, VTEX_FORMAT_R8, VTEX_FORMAT_BC4, VTEX_FORMAT_BC7:
			panic("Decode me")
		case VTEX_FORMAT_BGRA8888:
			// Swap red and blue
			bufferIndex := 0
			var a byte
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					a = (*buffer)[bufferIndex]
					(*buffer)[bufferIndex] = (*buffer)[bufferIndex + 2]
					(*buffer)[bufferIndex + 2] = a
					bufferIndex += 4
				}
			}
		case VTEX_FORMAT_PNG_R8G8B8A8_UINT, VTEX_FORMAT_R8G8B8A8_UINT:
			// Nothing to do
		default:
			panic("Unknown image format in decodeBuffer")
	}
}
