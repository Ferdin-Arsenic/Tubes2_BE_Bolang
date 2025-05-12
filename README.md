# <h1 align="center">Tugas Besar 2 IF2211 Strategi Algoritma</h1>
<h2 align="center">Semester II tahun 2024/2025</h2>
<h3 align="center">Pemanfaatan Algoritma BFS dan DFS dalam Pencarian Recipe pada Permainan</h3>

<p align="center">
  <img src="game.png" alt="Main" width="400">
</p>

## Table of Contents
- [Description](#description)
- [DFS, BFS, & Bidirectional Search](#algorithms-implemented)
- [Program Structure](#program-structure)
- [Requirements & Installation](#requirements--installation)
- [Author](#author)
- [References](#references)

## Description
This program is a web application for 

## Algorithms Implemented
### 1. BFS
### 2. DFS
### 3. Bidirectional

## Program Structure
### Backend
```
.
├── doc
│   └── Bolang.pdf
├── Dockerfile
├── README.md
└── src
    ├── bfs.go
    ├── bidirectional.go
    ├── data
    │   └── elements.json
    ├── dfs.go
    ├── go.mod
    ├── go.sum
    ├── main.go
    ├── scrapper.go
    └── treebuilder.go

4 directories, 12 files
```
- **src** : contains source code for algorithms and other backend implementations for the Web Application
- **doc** : contains the assignment report and program documentation.

### Frontend
```
.
├── docker-compose.yml
├── Dockerfile
├── eslint.config.mjs
├── next.config.ts
├── next-env.d.ts
├── package.json
├── package-lock.json
├── postcss.config.mjs
├── public
│   ├── file.svg
│   ├── globe.svg
│   ├── icon
│   │   └── search.svg
│   ├── image.png
│   ├── next.svg
│   ├── vercel.svg
│   └── window.svg
├── README.md
├── src
│   ├── app
│   │   ├── favicon.ico
│   │   ├── globals.css
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   └── result
│   │       └── page.tsx
│   ├── components
│   │   ├── ParameterBar
│   │   │   └── page.tsx
│   │   ├── ResultBar
│   │   │   └── page.tsx
│   │   ├── TreeClientWrapper.tsx
│   │   └── TreeRecipe
│   │       ├── page.tsx
│   │       └── TreeClient.tsx
│   ├── data
│   │   ├── basic.json
│   │   ├── elements.json
│   │   ├── images.json
│   │   └── recipe.json
│   ├── Tree
│   │   └── TreeRecipe.tsx
│   └── types
│       └── types.tsx
└── tsconfig.json

13 directories, 33 files
```


## Requirements & Installation
Before running the application, make sure the following dependencies are installed:
### Requirements:
### Installation and Running the Program:
1. Download and Install Dependencies:

## Author
| **NIM**  | **Nama Anggota**               | **Github** |
| -------- | ------------------------------ | ---------- |
| 13523025 | M. Rayhan Farrukh              | [grwna](https://github.com/grwna) |
| 13523115 | Azfa Radhiyya Hakim            | [azfaradhi](https://github.com/azfaradhi) | 
| 13523117 | Ferdin Arsenarendra Purtadi    | [Ferdin-Arsenic](https://github.com/Ferdin-Arsenic) |

## References
- [Spesifikasi Tugas Besar 2 Stima 2024/2025](https://docs.google.com/document/d/1aQB5USxfUCBfHmYjKl2wV5WdMBzDEyojE5yxvBO3pvc/edit?tab=t.0)
- [Slide Kuliah IF2211 2024/2025 Algoritma BFS dan DFS (Bagian 1)](https://informatika.stei.itb.ac.id/~rinaldi.munir/Stmik/2024-2025/13-BFS-DFS-(2025)-Bagian1.pdf)
- [Slide Kuliah IF2211 2024/2025 Algoritma BFS dan DFS (Bagian 2)](https://informatika.stei.itb.ac.id/~rinaldi.munir/Stmik/2024-2025/14-BFS-DFS-(2025)-Bagian2.pdf)
- [Little Alchemy 2 Fandom](https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2))
- [Golang Documentation](https://go.dev/doc/)