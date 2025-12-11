# Logo and Favicon Update Design Document

## Overview
This document outlines the design for updating logo and favicon assets across both the frontend and widget projects in the Botla platform. The update involves integrating newly generated assets with specific dimensions for various use cases.

## Current State Analysis

### Frontend Project
The frontend project currently contains the following assets in `frontend/public/`:
- apple-touch-icon.png (32.3KB)
- favicon.ico (3.2KB)
- favicon.png (1.6KB)
- logo-1024.png (1086.4KB)
- logo-128.png (16.5KB)
- logo-256.png (64.5KB)
- logo-512.png (252.6KB)
- og-image.png (811.6KB)
- vite.svg (1.5KB)

Current HTML head configuration references `/vite.svg` as the icon.

### Widget Project
The widget project currently contains the following assets in `widget/public/`:
- demo-dev.html (0.6KB)
- demo.html (0.4KB)
- favicon.ico (3.2KB)
- favicon.png (1.6KB)
- logo-128.png (16.5KB)
- logo-256.png (64.5KB)
- vite.svg (1.5KB)

Current HTML head configuration references `/vite.svg` as the icon.

## Asset Inventory

### Frontend Assets (frontend/public/)
| Filename | Dimensions | Purpose |
|----------|------------|---------|
| logo-1024.png | 1024x1024 | Base logo |
| logo-512.png | 512x512 | Large logo |
| logo-256.png | 256x256 | Medium logo |
| logo-128.png | 128x128 | Small logo |
| favicon.ico | 32x32 | Browser favicon |
| favicon.png | 32x32 | Modern favicon |
| apple-touch-icon.png | 180x180 | iOS home screen |
| og-image.png | 1200x630 | Social sharing |

### Widget Assets (widget/public/)
| Filename | Dimensions | Purpose |
|----------|------------|---------|
| logo-256.png | 256x256 | Medium logo |
| logo-128.png | 128x128 | Small logo |
| favicon.ico | 32x32 | Browser favicon |
| favicon.png | 32x32 | Modern favicon |

## Implementation Strategy

### Frontend Project Updates
1. Replace existing assets in `frontend/public/` with the new asset files
2. Update `frontend/index.html` to reference the new favicon files instead of `vite.svg`
3. Add apple touch icon declaration to `frontend/index.html`
4. Verify React component references to logo files (if any)
5. Confirm Open Graph meta tags reference the correct og-image.png

### Widget Project Updates
1. Replace existing assets in `widget/public/` with the new asset files
2. Update `widget/index.html` to reference the new favicon files instead of `vite.svg`
3. Validate asset references in widget HTML templates (`demo-dev.html`, `demo.html`)
4. Confirm favicon declarations in all widget HTML files

## Technical Considerations

### File Management
- Preserve existing filenames to maintain backward compatibility
- Ensure all new assets are optimized for web use
- Maintain consistent naming conventions across both projects

### Integration Points
- HTML head meta tags for favicon declarations
- React component references to logos
- CSS background image references (if any)
- Manifest files for progressive web app configuration (if applicable)

### HTML Meta Tag Updates

#### Frontend (`frontend/index.html`)
- Replace `<link rel="icon" type="image/svg+xml" href="/vite.svg" />` with proper favicon declarations
- Add apple touch icon link: `<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">`

#### Widget (`widget/index.html`)
- Replace `<link rel="icon" type="image/svg+xml" href="/vite.svg" />` with proper favicon declarations

## Validation Requirements

### Frontend
- Visual verification of logo rendering across different components
- Favicon display in browser tabs
- Apple touch icon appearance on iOS devices
- Social media preview rendering with og-image.png
- Verification that all HTML meta tags are correctly updated

### Widget
- Logo rendering in embedded chat widget
- Favicon display when widget is opened in standalone mode (if applicable)
- Verification that all HTML meta tags are correctly updated

## Rollout Plan
1. Backup existing assets in both projects
2. Replace asset files in `frontend/public/` and `widget/public/` with new versions
3. Update HTML meta tags in `frontend/index.html` and `widget/index.html`
4. Deploy updated code to staging environment
5. Conduct visual regression testing
6. Deploy to production after validation

## Success Criteria
- All new assets display correctly across supported browsers and devices
- No broken image links or missing assets
- Consistent branding presentation across frontend and widget
- Maintained performance standards (optimized asset sizes)- Consistent branding presentation across frontend and widget
