# GCP Go Microservices CI/CD Step-by-Step Guide

---

##  STEP-BY-STEP: CI/CD Go Microservices ke GCP (Artifact Registry + VM + GitHub Actions)

---

### 1. **GCP Preparation**

#### 1.1 **Buat Project GCP & Aktifkan Billing**
- Login ke Google Cloud Console.
- Buat **project baru** (`ex. third-project-467207` atau nama lain).
- Aktifkan **Billing** (wajib agar bisa pakai Artifact Registry, VM, dsb).

#### 1.2 **Buat VM Instance (Ubuntu)**
- **Compute Engine → VM Instances → Create Instance**
  - Pilih **OS: Ubuntu 24.04 LTS**.
  - Pilih **region: asia-southeast2 (Jakarta)**.
  - **Set Firewall**: Centang Allow HTTP/HTTPS *jika perlu*.
  - [*Lihat langkah SSH Key di bawah untuk pengisian SSH!*]

#### 1.3 **Generate SSH Key (Laptop Developer)**
- **Mac/Linux:**
  ```bash
  ssh-keygen -t rsa -b 4096 -C "ci-cd-tokokecil"
  cat ~/.ssh/id_rsa.pub     # Untuk copy public key
  cat ~/.ssh/id_rsa         # Untuk copy private key (diisi ke GitHub Secret)
  ```
- **Windows (Git Bash):**
  ```bash
  ssh-keygen -t rsa -b 4096 -C "ci-cd-tokokecil"
  cat ~/.ssh/id_rsa.pub     # Buka file ini atau pakai Notepad
  cat ~/.ssh/id_rsa         # Sama, copy isi file ini untuk ke Secret
  ```

#### 1.4 **Assign Public Key ke VM**
- Saat create VM: **paste** hasil `cat ~/.ssh/id_rsa.pub` ke kolom **SSH Keys**.
- Atau, **edit VM** → **SSH Keys** → paste public key.

#### 1.5 **Tes SSH dari Laptop ke VM**
```bash
ssh -i ~/.ssh/id_rsa <vm_username>@<VM_PUBLIC_IP>
```
Contoh:  
```bash
ssh -i ~/.ssh/id_rsa slwijaya@35.128.69.109
```
Pastikan login berhasil (tidak ada error).

#### 1.6 **Set Firewall Rules**
- **Compute Engine > VM instances > [klik nama VM] > View network details**
- Klik tab **Firewall rules** > **Create firewall rule**
  - **Direction**: Ingress
  - **Targets**: All instances in the network
  - **Source IP**: 0.0.0.0/0 (atau batasi sesuai kebutuhan)
  - **Protocols & Ports**:  
    - 22 (SSH)
    - 80 (HTTP, jika perlu)
    - 443 (HTTPS, jika perlu)
    - 8000, 8080, 8081 (port service Go-mu)
    - gRPC (tambahkan port, misal 50051 jika perlu)

#### 1.7 **Buat Artifact Registry**
- **Artifact Registry → Repositories → Create**
  - **Type**: Docker
  - **Region**: asia-southeast2
  - **Repository name**: (misal) `docker-asia-v2`
- Pastikan repository **ter-create** dan kosong.

#### 1.8 **Buat Service Account (SA)**
- **IAM & Admin > Service Accounts > Create Service Account**
  - Beri nama, misal `ci-cd-github-actions`
  - Setelah dibuat, klik service account → **GRANT ROLES**:
    - **Artifact Registry Writer**
    - **Compute Instance Admin (v1)**
    - **Viewer** *(opsional, untuk read-only akses)*
- **Buat kunci baru**:  
  - Klik service account > **KEYS** > **Add key** > **Create new key** (JSON)
  - Download dan simpan **JSON key** — ini yang nanti di-upload ke GitHub.

---

### 2. **Setup Secrets di GitHub**

Pada repo GitHub kamu:
- **Settings → Secrets and variables → Actions → New repository secret**
- Isi berikut:
  - `GCP_SA_KEY` → isi dengan isi **file JSON Service Account** (copy seluruh JSON).
  - `VM_HOST` → isi IP VM, misal: `34.128.69.109`
  - `VM_USER` → isi username VM, misal: `slwijaya`
  - `VM_SSH_KEY` → isi dengan **private key** (`id_rsa`, hasil `cat ~/.ssh/id_rsa`, **BUKAN** .pub!)
- **Pastikan tidak ada spasi/enter berlebihan.**

---

### 3. **Buat File Workflow: `.github/workflows/deploy.yml`**

Contoh **deploy.yml** (dengan penjelasan tiap langkah):

```yaml
name: Deploy Go Microservices to GCP VM

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest

    env:
      # GANTI semua value ini sesuai project GCP dan repo-mu!
      PROJECT_ID: <GCP_PROJECT_ID>                     # ex: third-project-467207
      REGION: <GCP_REGION>                             # ex: asia-southeast2
      REPOSITORY: <ARTIFACT_REGISTRY_REPO>             # ex: docker-asia-v2
      REGISTRY: ${{ env.REGION }}-docker.pkg.dev/${{ env.PROJECT_ID }}/${{ env.REPOSITORY }}
      VM_HOST: ${{ secrets.VM_HOST }}
      VM_USER: ${{ secrets.VM_USER }}
      VM_SSH_KEY: ${{ secrets.VM_SSH_KEY }}
      GCP_SA_KEY: ${{ secrets.GCP_SA_KEY }}

    steps:
    - name: Checkout source code
      uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Authenticate to Google Cloud
      uses: google-github-actions/auth@v2
      with:
        credentials_json: ${{ secrets.GCP_SA_KEY }}

    - name: Set up gcloud CLI
      uses: google-github-actions/setup-gcloud@v2
      with:
        project_id: ${{ env.PROJECT_ID }}

    # Gunakan hanya root registry agar Docker bisa auth ke semua repo project-mu
    - name: Configure Docker to use gcloud as a credential helper
      run: |
        gcloud auth configure-docker ${{ env.REGION }}-docker.pkg.dev --quiet

    # Build & Push Auth Service image
    - name: Build & Push Auth Service
      run: |
        docker buildx build --platform linux/amd64,linux/arm64 \
          -t ${{ env.REGISTRY }}/auth-service:latest ./<PATH_TO_AUTH_SERVICE> --push

    # Build & Push Gateway Service image
    - name: Build & Push Gateway Service
      run: |
        docker buildx build --platform linux/amd64,linux/arm64 \
          -t ${{ env.REGISTRY }}/gateway-service:latest ./<PATH_TO_GATEWAY_SERVICE> --push

    # Deploy to GCP VM via SSH, pull image terbaru dan up!
    - name: Deploy to GCP VM via SSH
      uses: appleboy/ssh-action@v1.0.0
      with:
        host: ${{ env.VM_HOST }}
        username: ${{ env.VM_USER }}
        key: ${{ env.VM_SSH_KEY }}
        script: |
          gcloud auth configure-docker ${{ env.REGION }}-docker.pkg.dev --quiet
          docker pull ${{ env.REGISTRY }}/auth-service:latest
          docker pull ${{ env.REGISTRY }}/gateway-service:latest
          docker rm -f auth-service || true
          docker rm -f gateway-service || true
          docker-compose up -d
          docker image prune -af || true
```

**Penjelasan:**
- Semua environment (`env:`) diisi dari **Secrets** repo.
- **`gcloud auth configure-docker`** WAJIB pakai _root registry_ (`asia-southeast2-docker.pkg.dev`).
- Buildx bisa multi-arch (`amd64, arm64`).
- SSH ke VM otomatis pull image terbaru, restart service, dan prune image lama.

---

### 4. **Summary Table — Perintah SSH Key di Mac/Win**

| Step         | Mac/Linux Command                     | Windows (Git Bash/PowerShell)   |
|--------------|---------------------------------------|---------------------------------|
| Generate key | `ssh-keygen -t rsa -b 4096 -C ...`    | Sama (pakai Git Bash)           |
| Copy pub     | `cat ~/.ssh/id_rsa.pub`               | `cat ~/.ssh/id_rsa.pub`         |
| Copy priv    | `cat ~/.ssh/id_rsa`                   | `cat ~/.ssh/id_rsa`             |
| SSH to VM    | `ssh -i ~/.ssh/id_rsa user@ip`        | Sama                            |

---

### 5. **Checklist Before Run (Pastikan Semua Siap!)**
- [x] **Project, Billing, VM, Artifact Registry, Service Account, Firewall, JSON Key, SSH VM ready**
- [x] **Repo sudah punya file `deploy.yml`**
- [x] **GitHub Secrets sudah benar (GCP_SA_KEY, VM_HOST, VM_USER, VM_SSH_KEY)**
- [x] **Test SSH ke VM dari laptop — pastikan connect**
- [x] **docker-compose.yml di repo sudah benar**

**Tinggal push ke main, workflow akan otomatis build-push-deploy!**

---

### 6. **Notes & Troubleshooting**
- Jika error 403 di buildx push: **99% salah konfigurasi `gcloud auth configure-docker`!**
- Pastikan role **Artifact Registry Writer** sudah di Service Account, dan Secrets tidak ada typo/whitespace.
- SSH Key yang dimasukkan ke Secret **HARUS yang private (`id_rsa`)**, BUKAN `.pub`.
- GCP Service Account JSON di Secret: **seluruh file** (bukan string/nama file saja).

---

> **Created by SF!**

