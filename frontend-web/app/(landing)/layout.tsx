import type { Metadata } from "next";

export const metadata: Metadata = {
	title: "Flock | Move Together, Find Your People",
	description:
		"Flock is a community discovery platform that helps individuals find local groups, in-person events, and online communities where they can build genuine, long-lasting relationships.",
};

export default function HomeLayout({
	children,
}: Readonly<{
	children: React.ReactNode;
}>) {
	return (
		<html lang="en" className="light">
			<head>
				<link href="https://fonts.googleapis.com/css2?family=Geist:wght@400;700;800&family=Geist+Mono:wght@700&family=Hanken+Grotesk:wght@400;500;700&family=JetBrains+Mono:wght@600&display=swap" rel="stylesheet" />
				<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap" rel="stylesheet" />
				<script src="https://cdn.tailwindcss.com?plugins=forms,container-queries"></script>
				<script dangerouslySetInnerHTML={{
					__html: `
					tailwind.config = {
						darkMode: "class",
						theme: {
							extend: {
								"colors": {
									"surface-container-lowest": "#ffffff",
									"on-primary-fixed": "#410000",
									"on-tertiary-fixed-variant": "#584329",
									"on-error": "#ffffff",
									"surface": "#fff8f1",
									"on-primary": "#ffffff",
									"inverse-on-surface": "#f7f0e7",
									"on-background": "#1e1b16",
									"surface-variant": "#e8e1d9",
									"secondary": "#3b6471",
									"primary-fixed-dim": "#ffb4a9",
									"primary-container": "#7f0303",
									"secondary-container": "#bce7f5",
									"tertiary-fixed-dim": "#e0c19f",
									"secondary-fixed-dim": "#a3cddb",
									"surface-container-low": "#faf3e9",
									"primary": "#570001",
									"background": "#fff8f1",
									"error": "#ba1a1a",
									"surface-container-high": "#eee7de",
									"on-error-container": "#93000a",
									"on-surface": "#1e1b16",
									"on-primary-container": "#ff8372",
									"surface-container": "#f4ede4",
									"tertiary": "#35240c",
									"primary-fixed": "#ffdad5",
									"surface-tint": "#b02d21",
									"surface-bright": "#fff8f1",
									"inverse-primary": "#ffb4a9",
									"on-secondary-container": "#3f6875",
									"surface-container-highest": "#e8e1d9",
									"inverse-surface": "#33302a",
									"on-tertiary-fixed": "#281803",
									"tertiary-container": "#4d3920",
									"on-secondary-fixed-variant": "#214c58",
									"on-secondary": "#ffffff",
									"on-tertiary": "#ffffff",
									"outline": "#8d706c",
									"on-secondary-fixed": "#001f27",
									"surface-dim": "#e0d9d0",
									"error-container": "#ffdad6",
									"secondary-fixed": "#bfe9f8",
									"on-surface-variant": "#59413d",
									"on-tertiary-container": "#bfa382",
									"outline-variant": "#e1bfb9",
									"on-primary-fixed-variant": "#8e120c",
									"tertiary-fixed": "#fdddb9"
								},
								"borderRadius": {
									"DEFAULT": "0px",
									"lg": "0px",
									"xl": "0px",
									"full": "0px"
								},
								"spacing": {
									"flat-shadow-offset": "6px",
									"margin": "32px",
									"gutter": "24px",
									"container-max": "1280px",
									"unit": "4px"
								},
								"fontFamily": {
									"body-lg": ["Hanken Grotesk"],
									"display-pixel": ["Geist Mono"],
									"headline-md": ["Geist"],
									"headline-lg-mobile": ["Geist"],
									"label-sm": ["JetBrains Mono"],
									"headline-lg": ["Geist"],
									"body-md": ["Hanken Grotesk"]
								},
								"fontSize": {
									"body-lg": ["18px", {"lineHeight": "28px", "fontWeight": "400"}],
									"display-pixel": ["84px", {"lineHeight": "84px", "letterSpacing": "-2px", "fontWeight": "700"}],
									"headline-md": ["24px", {"lineHeight": "32px", "fontWeight": "700"}],
									"headline-lg-mobile": ["32px", {"lineHeight": "36px", "fontWeight": "800"}],
									"label-sm": ["12px", {"lineHeight": "16px", "fontWeight": "600"}],
									"headline-lg": ["48px", {"lineHeight": "56px", "letterSpacing": "-1px", "fontWeight": "800"}],
									"body-md": ["16px", {"lineHeight": "24px", "fontWeight": "400"}]
								}
							},
						},
					}
					`
				}} />
			</head>
			<body>
				{children}
			</body>
		</html>
	);
}
