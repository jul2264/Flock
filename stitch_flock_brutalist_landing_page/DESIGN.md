---
name: Flock
colors:
  surface: '#fff8f1'
  surface-dim: '#e0d9d0'
  surface-bright: '#fff8f1'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#faf3e9'
  surface-container: '#f4ede4'
  surface-container-high: '#eee7de'
  surface-container-highest: '#e8e1d9'
  on-surface: '#1e1b16'
  on-surface-variant: '#59413d'
  inverse-surface: '#33302a'
  inverse-on-surface: '#f7f0e7'
  outline: '#8d706c'
  outline-variant: '#e1bfb9'
  surface-tint: '#b02d21'
  primary: '#570001'
  on-primary: '#ffffff'
  primary-container: '#7f0303'
  on-primary-container: '#ff8372'
  inverse-primary: '#ffb4a9'
  secondary: '#3b6471'
  on-secondary: '#ffffff'
  secondary-container: '#bce7f5'
  on-secondary-container: '#3f6875'
  tertiary: '#35240c'
  on-tertiary: '#ffffff'
  tertiary-container: '#4d3920'
  on-tertiary-container: '#bfa382'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#ffdad5'
  primary-fixed-dim: '#ffb4a9'
  on-primary-fixed: '#410000'
  on-primary-fixed-variant: '#8e120c'
  secondary-fixed: '#bfe9f8'
  secondary-fixed-dim: '#a3cddb'
  on-secondary-fixed: '#001f27'
  on-secondary-fixed-variant: '#214c58'
  tertiary-fixed: '#fdddb9'
  tertiary-fixed-dim: '#e0c19f'
  on-tertiary-fixed: '#281803'
  on-tertiary-fixed-variant: '#584329'
  background: '#fff8f1'
  on-background: '#1e1b16'
  surface-variant: '#e8e1d9'
typography:
  display-pixel:
    fontFamily: Geist Mono
    fontSize: 84px
    fontWeight: '700'
    lineHeight: 84px
    letterSpacing: -2px
  headline-lg:
    fontFamily: Geist
    fontSize: 48px
    fontWeight: '800'
    lineHeight: 56px
    letterSpacing: -1px
  headline-lg-mobile:
    fontFamily: Geist
    fontSize: 32px
    fontWeight: '800'
    lineHeight: 36px
  headline-md:
    fontFamily: Geist
    fontSize: 24px
    fontWeight: '700'
    lineHeight: 32px
  body-lg:
    fontFamily: Hanken Grotesk
    fontSize: 18px
    fontWeight: '400'
    lineHeight: 28px
  body-md:
    fontFamily: Hanken Grotesk
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  label-sm:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: '600'
    lineHeight: 16px
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  unit: 4px
  gutter: 24px
  margin: 32px
  container-max: 1280px
  flat-shadow-offset: 6px
---

## Brand & Style
The design system is rooted in **Neo-Brutalism**, a style that rejects standard "polite" UI conventions in favor of raw, structural honesty. It is designed for users who appreciate high-impact visuals and a clear sense of hierarchy. 

The aesthetic is defined by:
- **Structural Integrity:** Elements are treated as physical blocks with thick, consistent black borders.
- **Bento Box Layouts:** Content is organized into modular containers, creating a predictable yet dynamic grid.
- **Flat Shadows:** Depth is achieved not through blurs, but through hard-edged, offset shadows that emphasize the 2D plane.
- **High Contrast:** The intersection of a sophisticated heritage palette with aggressive black linework creates a tension that is both modern and timeless.

## Colors
The palette bridges classic, heritage tones with high-contrast brutalist strokes. 

- **Primary (Maroon #7F0303):** Used for critical actions, key branding moments, and primary headlines.
- **Secondary (Light Blue #96C0CE):** Acts as a refreshing counterpoint to the heavier tones, used for supporting elements and interactive states.
- **Tertiary (Tan #D8BA98):** Provides a warm, organic feel to structural elements.
- **Neutral (Alabaster #EFE8DF):** The primary background color, reducing the harshness of pure white while maintaining high legibility.
- **Midnight Blue (#0F414A):** Utilized for body text and deep contrast areas where black might feel too flat.
- **Accent (Black #000000):** Reserved exclusively for borders and hard shadows to define the "Brutalist" structure.

## Typography
Typography is a structural element in this design system. 

- **Brand & Display:** The logo and hero headers use **Geist Pixel** (rendered via Geist Mono for technical precision) to evoke a digital-first, modular feel.
- **Headlines:** **Geist** provides a clean, geometric sans-serif look that maintains authority in bold and extra-bold weights.
- **Body:** **Hanken Grotesk** is chosen for its exceptional readability and contemporary professional tone.
- **Labels:** **JetBrains Mono** is used for utility text, metadata, and labels to reinforce the technical, brutalist aesthetic.

All headlines should be set with tight letter-spacing to enhance the "blocky" feel of the layout.

## Layout & Spacing
This design system utilizes a **Fixed Bento Grid** model. Elements are housed in distinct cells that align to a 12-column grid on desktop and a 2-column grid on mobile.

- **The Bento Box:** Each component is a "cell" with a mandatory 2px or 3px black border.
- **Spacing Rhythm:** Based on a 4px baseline. Gutters are generous (24px) to allow the heavy borders "room to breathe" without feeling cluttered.
- **Safe Areas:** Mobile devices use a 16px margin, while desktop scales to 32px or larger depending on screen width.
- **Shadow Offset:** Depth is created by a `flat-shadow-offset`. Interactive elements should physically shift on the X and Y axis when clicked to "meet" their shadow.

## Elevation & Depth
Elevation is not conveyed through light sources or blurs. Instead, we use **Hard Layering**:

1.  **Level 0 (Base):** The Alabaster background (#EFE8DF).
2.  **Level 1 (Cards/Containers):** White or Tan surfaces with a 2px black border.
3.  **Level 2 (Interactive):** Elements with a hard, unblurred shadow (Color: Black, Opacity: 100%) offset by 6px down and 6px right.
4.  **State Changes:** On hover, the shadow offset may increase. On press, the element translates (+4px, +4px) to simulate a physical push-down effect.

## Shapes
While traditional Brutalism uses sharp 0px corners, this design system adopts a **Soft-Brutalist** approach to improve UX and approachability.

- **Standard Radius:** 0.25rem (4px). This subtle rounding takes the "edge" off the thick black borders while maintaining a rigid, masculine structure.
- **Pills:** Used exclusively for tags or chips to provide visual variety against the strictly rectangular bento grid.
- **Borders:** All primary containers must have a solid black border (#000000) with a minimum width of 2px.

## Components
- **Buttons:** High-contrast blocks. Primary buttons use Maroon (#7F0303) with white text and a thick black shadow. Secondary buttons use Light Blue (#96C0CE).
- **Cards (Bento Cells):** Must have a 2px black border. Backgrounds can vary between White, Tan, or Alabaster to differentiate content types.
- **Input Fields:** Flat white background, 2px black border, and JetBrains Mono for placeholder text. No inner shadows; use a solid color change on focus.
- **Chips:** Pill-shaped with a 1px black border. Use these for categories within the bento boxes.
- **Lists:** Separated by horizontal 2px black lines. No rounded corners on list items to maintain the vertical "stack" look.
- **Checkboxes:** Square with sharp corners, 2px border. Use the Primary Maroon for the checked state.