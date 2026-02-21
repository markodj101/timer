# Wails3 Timer (Conference Edition)

Ovo je moćan alat za upravljanje vremenom namijenjen kongresima i događajima. Aplikacija omogućava potpunu kontrolu tajmera na glavnom monitoru, dok se prikaz za publiku emituje na sekundarnom ekranu.

*(Zamijeni putanju do svog screenshot-a)*

## Glavne Funkcionalnosti

* **Svelte + Wails3:** Ultra brzi frontend uparen sa stabilnim Go backendom.
* **Multi-Screen Support:** Automatska detekcija i podrška za projektore/eksterne monitore (Public View).
* **Customization:** * Prilagođavanje teksta i vremena.
* **Custom Sound (IPC):** Podrška za prilagođene zvukove upozorenja putem IPC komunikacije.


* **Cross-Platform:** Potpuna podrška za **Linux** i **Windows 11**.

---

## Instalacija i Pokretanje

### Preduslovi

* **Go** 1.25.x ili noviji
* **Node.js** v22+ i **npm**
* **Wails v3 Alpha** (`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`)
* *(Samo za Linux)*: `libwebkit2gtk-4.1-dev`

### Razvoj (Development)

Da pokreneš aplikaciju sa "hot-reload" funkcijom (izmjene se vide odmah):

```bash
wails3 dev

```

### Produkcija (Build)

Za kreiranje izvršnog fajla u `build` direktorijumu:

```bash
wails3 build

```

---

## Struktura Projekta

* `frontend/` - Svelte aplikacija (UI/UX).
* `main.go` - Glavni ulaz u Go backend i konfiguracija prozora.
* `app.go` - Backend logika, servisi za čitanje fajlova i zvukova.
* `frontend/dist` - Lokacija kompajliranog frontenda (embedovan u binary).

---

## Podrška za ekrane (Linux/Win)

Aplikacija je optimizovana za rad sa više monitora:

1. **Kontrolni panel:** Otvara se na primarnom monitoru.
2. **Display View:** Automatski traži sekundarni monitor i širi se preko cijelog ekrana (Fullscreen).

---

### To Do:
