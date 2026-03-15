# 🐑 Herdbuch - Sheep Herd Management System

A web-based application for tracking and managing sheep herd activities, specifically designed for Ferdinand, Wolke, Sunflower, Flocke, and Hope.

## ✨ Features

- **Individual Sheep Tracking** - Log entries for specific sheep or the entire herd
- **Preset Messages** - Quick entry buttons for common activities:
  - 🌾 Fütterung (Feeding)
  - 🩺 Gesundheitscheck (Health Check)
  - ✂️ Schur (Shearing)
- **Entry History** - View all historical entries with timestamps
- **Scope Filtering** - Filter entries by individual sheep or view all
- **Responsive Design** - Works on desktop and mobile devices
- **Progressive Web App** - Can be installed on devices like a native app
- **Cute Sheep Icons** - Custom-designed sheep-themed icons and interface

## 🚀 Getting Started

### Prerequisites
- Go 1.19 or higher
- Modern web browser

### Installation & Running

1. **Clone the repository**
   ```bash
   git clone https://github.com/codeispoetry/herdbook.git
   cd herdbook
   ```

2. **Install Go dependencies**
   ```bash
   go mod tidy
   ```

3. **Start the server**
   ```bash
   go run main.go
   ```

4. **Access the application**
   Open your browser to: `http://localhost:9002`

## 🐑 Sheep Management

The app tracks activities for:
- **Ferdinand** - Individual sheep tracking
- **Wolke** - Individual sheep tracking  
- **Sunflower** - Individual sheep tracking
- **Flocke** - Individual sheep tracking
- **Hope** - Individual sheep tracking
- **Herde** - Entries for the entire herd

## 🛠 Technology Stack

- **Backend**: Go with SQLite3 database
- **Frontend**: HTML5, CSS3, Vanilla JavaScript
- **Database**: SQLite3 for local data storage
- **Icons**: Custom SVG sheep-themed icons
- **PWA**: Service Worker for offline capabilities

## 📝 API Endpoints

- `GET /` - Serve the web interface
- `POST /post` - Add new herd book entry
- `GET /list` - Retrieve all entries

### POST /post
```json
{
  "scope": "Ferdinand",
  "message": "Fütterung durchgeführt"
}
```

### GET /list Response
```json
[
  {
    "id": 1,
    "timestamp": "2026-03-15T10:30:00Z",
    "scope": "Ferdinand",
    "message": "Fütterung durchgeführt"
  }
]
```

## 🎨 Interface Features

- **Sheep Theme** - Blue sky colors with fluffy white sheep icons
- **German Interface** - Designed for German-speaking users
- **Filter System** - Quick filtering by sheep category
- **Auto-refresh** - Entries update automatically after submission
- **Responsive Layout** - Mobile-friendly design

## 📱 Progressive Web App

The application can be installed as a PWA on mobile devices and desktop for offline access and native app-like experience.

## 🗃 Database Schema

```sql
CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    scope TEXT NOT NULL,
    message TEXT NOT NULL
);
```

## 🤝 Contributing

Feel free to submit issues and enhancement requests!

## 📄 License

This project is created for managing sheep herd activities.
