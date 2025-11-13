package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"git.miem.hse.ru/kg25-26/aisavelev.git/application/models"
	"git.miem.hse.ru/kg25-26/aisavelev.git/application/utils"

	"github.com/gorilla/mux"
)

const galleryTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Color Filter Gallery</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: Arial, sans-serif; padding: 20px; background: #f5f5f5; }
        .header { text-align: center; margin-bottom: 30px; padding: 20px; background: white; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .target-color { display: inline-block; width: 50px; height: 50px; border-radius: 50%; margin: 0 10px; border: 2px solid #333; }
        .gallery-container { display: flex; gap: 20px; max-width: 1400px; margin: 0 auto; }
        .column { flex: 1; display: flex; flex-direction: column; gap: 20px; }
        .image-card { background: white; border-radius: 10px; overflow: hidden; box-shadow: 0 4px 15px rgba(0,0,0,0.1); transition: transform 0.3s ease; }
        .image-card:hover { transform: translateY(-5px); }
		.image-container { width: 100%; display: flex; align-items: center; justify-content: center; background: transparent; padding: 10px; }
		.image-container img { max-width: 400px; width: auto; height: auto; display: block; }
        .image-info { padding: 15px; text-align: center; }
        .deviation { font-weight: bold; color: #666; margin-top: 5px; }
        .color-wheel { width: 400px; height: 400px; margin: 40px auto; position: relative; border-radius: 50%; background: conic-gradient(from 0deg, #ff0000, #ffff00, #00ff00, #00ffff, #0000ff, #ff00ff, #ff0000);border: 3px solid #333;box-shadow: 0 0 0 5px white, 0 0 0 8px #ccc; }
        .color-point { position: absolute; width: 10px; height: 10px; background: white; border: 2px solid #333; border-radius: 50%; transform: translate(-50%, -50%); cursor: pointer; }
        .color-label { position: absolute; background: white; padding: 5px 10px; border-radius: 15px; font-size: 12px; box-shadow: 0 2px 5px rgba(0,0,0,0.2); white-space: nowrap; cursor: pointer; }
        @media (max-width: 768px) { .gallery-container { flex-direction: column; } .color-wheel { width: 300px; height: 300px; } }
        .modal { display: none; position: fixed; z-index: 1000; left: 0; top: 0; width: 100%; height: 100%; background-color: rgba(0,0,0,0.9); }
        .modal-content { margin: auto; display: block; max-width: 90%; max-height: 90%; margin-top: 2%; }
        .close { position: absolute; top: 15px; right: 35px; color: #f1f1f1; font-size: 40px; font-weight: bold; cursor: pointer; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Color Filter Gallery</h1>
        <p>Target Color: 
            <span class="target-color" style="background: rgb({{.TargetColor.R}}, {{.TargetColor.G}}, {{.TargetColor.B}})"></span>
            | Tolerance: {{printf "%.1f" .Tolerance}}°
        </p>
    </div>
    <div class="gallery-container">
        <div class="column">{{range .Column1}}
            <div class="image-card">
                <div class="image-container">
                    <img src="/static/uploads/{{.SessionID}}/{{.FileName}}" alt="{{.FileName}}" onclick="openModal(this)">
                </div>
                <div class="image-info">
                    <div class="filename">{{.FileName}}</div>
                    <div class="deviation">Deviation: {{printf "%.2f" .Deviation}}°</div>
                </div>
            </div>{{end}}
        </div>
        <div class="column">{{range .Column2}}
            <div class="image-card">
                <div class="image-container">
                    <img src="/static/uploads/{{.SessionID}}/{{.FileName}}" alt="{{.FileName}}" onclick="openModal(this)">
                </div>
                <div class="image-info">
                    <div class="filename">{{.FileName}}</div>
                    <div class="deviation">Deviation: {{printf "%.2f" .Deviation}}°</div>
                </div>
            </div>{{end}}
        </div>
    </div>
    <div class="color-wheel" id="colorWheel"></div>
    <div id="imageModal" class="modal">
        <span class="close" onclick="closeModal()">&times;</span>
        <img class="modal-content" id="modalImage">
    </div>
    <script>
        function openModal(img) { 
            document.getElementById('imageModal').style.display = 'block'; 
            document.getElementById('modalImage').src = img.src; 
        }
        function closeModal() { 
            document.getElementById('imageModal').style.display = 'none'; 
        }
        
        const colorWheel = document.getElementById('colorWheel');
        const images = [ 
            {{range .Column1}} 
            { name: "{{.FileName}}", hue: {{printf "%.1f" .Hue}}, deviation: {{printf "%.2f" .Deviation}} }, 
            {{end}} 
            {{range .Column2}} 
            { name: "{{.FileName}}", hue: {{printf "%.1f" .Hue}}, deviation: {{printf "%.2f" .Deviation}} }, 
            {{end}} 
        ];
        
        // ПРОСТОЙ ТЕСТ - создадим 6 точек основных цветов
        const testHues = [0, 60, 120, 180, 240, 300]; // Красный, желтый, зеленый, голубой, синий, пурпурный
        
        testHues.forEach((hue, index) => {
            // Преобразуем hue в угол на круге
            const angle = ((hue - 90) * Math.PI) / 180; // -90 чтобы красный был сверху
            const radius = 160;
            const center = 200;
            const x = center + radius * Math.cos(angle);
            const y = center + radius * Math.sin(angle);
            
            const point = document.createElement('div');
            point.innerHTML = '•'; // Простой маркер
            point.style.position = 'absolute';
            point.style.left = x + 'px';
            point.style.top = y + 'px';
            point.style.color = 'white';
            point.style.fontSize = '24px';
            point.style.fontWeight = 'bold';
            point.style.textShadow = '2px 2px 4px black';
            point.style.transform = 'translate(-50%, -50%)';
            point.style.zIndex = '10';
            point.title = 'Hue: ' + hue + '°';
            
            colorWheel.appendChild(point);
            
            console.log('Test point', hue + '°', 'at', x, y);
        });
        
        // Теперь добавляем реальные точки изображений
        images.forEach((image) => {
            const angle = ((image.hue - 90) * Math.PI) / 180;
            const radius = 140; // Внутренний круг для реальных точек
            const center = 200;
            const x = center + radius * Math.cos(angle);
            const y = center + radius * Math.sin(angle);
            
            const point = document.createElement('div'); 
            point.className = 'color-point'; 
            point.style.left = x + 'px'; 
            point.style.top = y + 'px';
            point.style.backgroundColor = 'black'; // Черные точки для контраста
            point.onclick = function() { 
                const imgElement = document.querySelector('img[alt="' + image.name + '"]'); 
                if (imgElement) { 
                    imgElement.scrollIntoView({ behavior: 'smooth', block: 'center' }); 
                    imgElement.parentElement.parentElement.style.animation = 'highlight 1s ease'; 
                    setTimeout(function() { imgElement.parentElement.parentElement.style.animation = ''; }, 1000); 
                } 
            };
            
            const label = document.createElement('div'); 
            label.className = 'color-label'; 
            label.textContent = image.name; 
            label.style.left = (x + 20) + 'px'; 
            label.style.top = (y - 10) + 'px';
            label.onclick = function() { 
                const imgElement = document.querySelector('img[alt="' + image.name + '"]'); 
                if (imgElement) openModal(imgElement); 
            };
            
            colorWheel.appendChild(point); 
            colorWheel.appendChild(label);
            
            console.log('Image point:', image.name, 'hue:', image.hue, 'at', x, y);
        });
        
        const style = document.createElement('style'); 
        style.textContent = '@keyframes highlight { 0% { box-shadow: 0 4px 15px rgba(0,0,0,0.1); } 50% { box-shadow: 0 4px 20px rgba(255,0,0,0.5); } 100% { box-shadow: 0 4px 15px rgba(0,0,0,0.1); } }'; 
        document.head.appendChild(style);
        
        // Проверка видимости круга
        console.log('Color wheel visible:', colorWheel.offsetWidth > 0 && colorWheel.offsetHeight > 0);
        console.log('Color wheel position:', colorWheel.getBoundingClientRect());
    </script>
</body>
</html>`

func SetupRoutes(r *mux.Router) {
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/upload", UploadHandler)
	r.HandleFunc("/set-target", SetTargetHandler)
	r.HandleFunc("/set-target-by-filename", SetTargetByFilenameHandler)
	r.HandleFunc("/filter", FilterHandler)
	r.HandleFunc("/generate-html", GenerateHTMLHandler)
	r.HandleFunc("/preview", PreviewHandler)
	r.HandleFunc("/download", DownloadHandler)
	r.HandleFunc("/get-uploaded-images", GetUploadedImagesHandler)
	r.HandleFunc("/delete-image", DeleteImageHandler)
	r.HandleFunc("/debug-files", DebugFilesHandler)
	r.HandleFunc("/get-session-id", GetSessionIDHandler)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)
	SetSessionCookie(w, session.ID)

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := GetSession(r)
	SetSessionCookie(w, session.ID)

	log.Printf("Upload request for session: %s", session.ID)

	err := r.ParseMultipartForm(100 << 20) // 100MB
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Создание директории для загрузок
	uploadDir := filepath.Join("static", "uploads", session.ID)
	err = os.MkdirAll(uploadDir, 0755)
	if err != nil {
		log.Printf("Error creating upload directory: %v", err)
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	log.Printf("Upload directory: %s", uploadDir)

	files := r.MultipartForm.File["images"]
	var results []models.ImageAnalysis
	var errors []string

	log.Printf("Received %d files for upload", len(files))

	for _, fileHeader := range files {
		log.Printf("Processing file: %s, Size: %d bytes", fileHeader.Filename, fileHeader.Size)

		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening file %s: %v", fileHeader.Filename, err)
			errors = append(errors, "Cannot open file: "+fileHeader.Filename)
			continue
		}

		// Чтение данных файла
		fileData, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			log.Printf("Error reading file data %s: %v", fileHeader.Filename, err)
			errors = append(errors, "Cannot read file: "+fileHeader.Filename)
			continue
		}

		// Сохранение файла на диск
		filePath := filepath.Join(uploadDir, fileHeader.Filename)
		err = os.WriteFile(filePath, fileData, 0644)
		if err != nil {
			log.Printf("Error saving file %s: %v", fileHeader.Filename, err)
			errors = append(errors, "Cannot save file: "+fileHeader.Filename)
			continue
		}

		// Проверка
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Error verifying file %s: %v", filePath, err)
			errors = append(errors, "Cannot verify file: "+fileHeader.Filename)
			continue
		}

		log.Printf("File saved successfully: %s, Size: %d bytes", filePath, fileInfo.Size())

		// Анализ изображения
		analysis, err := analyzeImageFromData(fileData, fileHeader.Filename)
		if err != nil {
			log.Printf("Error analyzing image %s: %v", fileHeader.Filename, err)
			errors = append(errors, "Cannot analyze image: "+fileHeader.Filename)
			continue
		}

		results = append(results, analysis)
		log.Printf("Successfully processed: %s", fileHeader.Filename)
	}

	// Добавление результатов в сессию
	if len(results) > 0 {
		sessionsMux.Lock()
		session.Images = append(session.Images, results...)
		sessionsMux.Unlock()
	}

	// Отправка ответа
	response := map[string]interface{}{
		"success": len(results),
		"failed":  len(errors),
		"errors":  errors,
		"images":  results,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func analyzeImageFromData(data []byte, filename string) (models.ImageAnalysis, error) {
	reader := bytes.NewReader(data)

	img, format, err := image.Decode(reader)
	if err != nil {
		return models.ImageAnalysis{}, fmt.Errorf("failed to decode image %s (format detection): %v", filename, err)
	}

	log.Printf("Image %s decoded successfully: format=%s, bounds=%v", filename, format, img.Bounds())

	var rgbaImg *image.RGBA
	switch v := img.(type) {
	case *image.RGBA:
		rgbaImg = v
	case *image.NRGBA:
		rgbaImg = image.NewRGBA(v.Bounds())
		draw.Draw(rgbaImg, rgbaImg.Bounds(), v, v.Bounds().Min, draw.Src)
	default:
		b := img.Bounds()
		rgbaImg = image.NewRGBA(b)
		draw.Draw(rgbaImg, b, img, b.Min, draw.Src)
	}

	dominantColor := utils.CalculateDominantColor(rgbaImg)
	hsv := utils.RGBToHSV(dominantColor.R, dominantColor.G, dominantColor.B)

	analysis := models.ImageAnalysis{
		FileName:    filename,
		DominantRGB: dominantColor,
		HSV:         hsv,
		IsTarget:    false,
	}

	log.Printf("Image analysis for %s: RGB(%d,%d,%d), HSV(%.1f,%.2f,%.2f)",
		filename, dominantColor.R, dominantColor.G, dominantColor.B, hsv[0], hsv[1], hsv[2])

	return analysis, nil
}

func GetUploadedImagesHandler(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)

	sessionsMux.RLock()
	images := make([]models.ImageAnalysis, len(session.Images))
	copy(images, session.Images)
	sessionsMux.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

func SetTargetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := GetSession(r)
	SetSessionCookie(w, session.ID)

	file, _, err := r.FormFile("targetImage")
	if err != nil {
		rStr := r.FormValue("r")
		gStr := r.FormValue("g")
		bStr := r.FormValue("b")

		if rStr != "" && gStr != "" && bStr != "" {
			r, _ := strconv.Atoi(rStr)
			g, _ := strconv.Atoi(gStr)
			b, _ := strconv.Atoi(bStr)
			SetTargetColor(models.ColorRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255})
		} else {
			http.Error(w, "No target image provided", http.StatusBadRequest)
			return
		}
	} else {
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		SetTargetColor(utils.CalculateDominantColor(img))
	}

	target := GetTargetColor()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"r":   target.R,
		"g":   target.G,
		"b":   target.B,
		"hsv": utils.RGBToHSV(target.R, target.G, target.B),
	})
}

func SetTargetByFilenameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := GetSession(r)
	SetSessionCookie(w, session.ID)

	filename := r.FormValue("filename")

	sessionsMux.Lock()
	defer sessionsMux.Unlock()

	for i, img := range session.Images {
		if img.FileName == filename {
			for j := range session.Images {
				session.Images[j].IsTarget = false
			}
			session.Images[i].IsTarget = true
			SetTargetColor(img.DominantRGB)

			target := GetTargetColor()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"color": map[string]interface{}{
					"r":   target.R,
					"g":   target.G,
					"b":   target.B,
					"hsv": utils.RGBToHSV(target.R, target.G, target.B),
				},
			})
			return
		}
	}

	http.Error(w, "Image not found", http.StatusNotFound)
}

func DeleteImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := GetSession(r)
	SetSessionCookie(w, session.ID)

	filename := r.FormValue("filename")

	sessionsMux.Lock()
	defer sessionsMux.Unlock()

	for i, img := range session.Images {
		if img.FileName == filename {
			os.Remove(filepath.Join("static", "uploads", session.ID, filename))
			session.Images = append(session.Images[:i], session.Images[i+1:]...)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
			return
		}
	}

	http.Error(w, "Image not found", http.StatusNotFound)
}

func FilterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := GetSession(r)
	SetSessionCookie(w, session.ID)

	tolerance, err := strconv.ParseFloat(r.FormValue("tolerance"), 64)
	if err != nil {
		http.Error(w, "Invalid tolerance", http.StatusBadRequest)
		return
	}

	sortDirection := r.FormValue("sortDirection")
	if sortDirection == "" {
		sortDirection = "clockwise"
	}

	sessionsMux.RLock()
	defer sessionsMux.RUnlock()

	target := GetTargetColor()
	targetHSV := utils.RGBToHSV(target.R, target.G, target.B)
	var filtered []models.ImageAnalysis

	minSaturation := 0.1
	minValue := 0.05

	for _, analysis := range session.Images {
		hueDistance := math.Abs(analysis.HSV[0] - targetHSV[0])
		if hueDistance > 180 {
			hueDistance = 360 - hueDistance
		}

		saturation := analysis.HSV[1]
		value := analysis.HSV[2]

		if hueDistance <= tolerance &&
			saturation >= minSaturation &&
			value >= minValue {
			analysis.Deviation = hueDistance
			filtered = append(filtered, analysis)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if sortDirection == "clockwise" {
			return filtered[i].Deviation < filtered[j].Deviation
		}
		return filtered[i].Deviation > filtered[j].Deviation
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

func GenerateHTMLHandler(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)
	SetSessionCookie(w, session.ID)

	tolerance, _ := strconv.ParseFloat(r.FormValue("tolerance"), 64)
	sortDirection := r.FormValue("sortDirection")

	sessionsMux.RLock()
	defer sessionsMux.RUnlock()

	target := GetTargetColor()
	targetHSV := utils.RGBToHSV(target.R, target.G, target.B)
	var filtered []models.ImageAnalysis

	minSaturation := 0.1
	minValue := 0.05

	for _, analysis := range session.Images {
		hueDistance := math.Abs(analysis.HSV[0] - targetHSV[0])
		if hueDistance > 180 {
			hueDistance = 360 - hueDistance
		}

		if hueDistance <= tolerance &&
			analysis.HSV[1] >= minSaturation &&
			analysis.HSV[2] >= minValue {
			analysis.Deviation = hueDistance
			filtered = append(filtered, analysis)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if sortDirection == "clockwise" {
			return filtered[i].Deviation < filtered[j].Deviation
		}
		return filtered[i].Deviation > filtered[j].Deviation
	})

	mid := len(filtered) / 2
	var column1, column2 []models.ImageAnalysis
	if len(filtered) > 0 {
		column1 = filtered[:mid]
		column2 = filtered[mid:]
	}

	if sortDirection == "counterclockwise" {
		column1, column2 = column2, column1
		utils.ReverseSlice(column1)
	}

	type GalleryImage struct {
		models.ImageAnalysis
		SessionID string
		Hue       float64
	}

	galleryColumn1 := make([]GalleryImage, len(column1))
	galleryColumn2 := make([]GalleryImage, len(column2))

	for i, img := range column1 {
		galleryColumn1[i] = GalleryImage{
			ImageAnalysis: img,
			SessionID:     session.ID,
			Hue:           img.HSV[0],
		}
	}

	for i, img := range column2 {
		galleryColumn2[i] = GalleryImage{
			ImageAnalysis: img,
			SessionID:     session.ID,
			Hue:           img.HSV[0],
		}
	}

	data := struct {
		SessionID   string
		TargetColor models.ColorRGBA
		Tolerance   float64
		Column1     []GalleryImage
		Column2     []GalleryImage
		TargetHSV   [3]float64
	}{
		SessionID:   session.ID,
		TargetColor: target,
		Tolerance:   tolerance,
		Column1:     galleryColumn1,
		Column2:     galleryColumn2,
		TargetHSV:   targetHSV,
	}

	tmpl, err := template.New("gallery").Parse(galleryTemplate)
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func PreviewHandler(w http.ResponseWriter, r *http.Request) {
	GenerateHTMLHandler(w, r)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Disposition", "attachment; filename=color_gallery.html")
	GenerateHTMLHandler(w, r)
}

func DebugFilesHandler(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)
	uploadDir := filepath.Join("static", "uploads", session.ID)

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading directory: %v", err), http.StatusInternalServerError)
		return
	}

	var fileInfos []map[string]interface{}
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(uploadDir, file.Name())
		fileInfos = append(fileInfos, map[string]interface{}{
			"name": file.Name(),
			"size": info.Size(),
			"path": filePath,
			"url":  "/static/uploads/" + session.ID + "/" + file.Name(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileInfos)
}

func GetSessionIDHandler(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)

	response := map[string]interface{}{
		"session_id": session.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
