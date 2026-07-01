"use client";

import { useEffect, useRef } from "react";

export default function HomePage() {
	const slideshowRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		// Scroll animation triggers for neo-brutalist blocks
		const observerOptions = {
			threshold: 0.1
		};

		const observer = new IntersectionObserver((entries) => {
			entries.forEach(entry => {
				if (entry.isIntersecting) {
					(entry.target as HTMLElement).style.opacity = "1";
					(entry.target as HTMLElement).style.transform = "translateY(0)";
				}
			});
		}, observerOptions);

		document.querySelectorAll('.bento-card').forEach(card => {
			const el = card as HTMLElement;
			el.style.opacity = "0";
			el.style.transform = "translateY(20px)";
			el.style.transition = "all 0.6s cubic-bezier(0.23, 1, 0.32, 1)";
			observer.observe(el);
		});

		return () => observer.disconnect();
	}, []);

	const scrollSlideshow = (direction: number) => {
		slideshowRef.current?.scrollBy({ left: 320 * direction, behavior: 'smooth' });
	};

	return (
		<>
			<style dangerouslySetInnerHTML={{
				__html: `
					body {
						background-color: #fff8f1;
						color: #1e1b16;
						overflow-x: hidden;
					}
					.neo-shadow {
						box-shadow: 6px 6px 0px 0px rgba(0,0,0,1);
					}
					.neo-shadow-sm {
						box-shadow: 3px 3px 0px 0px rgba(0,0,0,1);
					}
					.neo-shadow-lg {
						box-shadow: 10px 10px 0px 0px rgba(0,0,0,1);
					}
					.active-push:active {
						transform: translate(4px, 4px);
						box-shadow: 2px 2px 0px 0px rgba(0,0,0,1);
					}
					.material-symbols-outlined {
						font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
						vertical-align: middle;
					}
					/* Custom hide scrollbar */
					.no-scrollbar::-webkit-scrollbar { display: none; }
					.no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }

					.bento-card {
						transition: all 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275);
					}
					.bento-card:hover {
						transform: translate(-2px, -2px);
						box-shadow: 8px 8px 0px 0px rgba(0,0,0,1);
					}
				`
			}} />
			
			<div className="font-body-md bg-background">
				{/* TopNavBar */}
				<header className="sticky top-0 z-50 bg-surface-container-lowest border-b-2 border-on-background">
					<nav className="flex justify-between items-center w-full px-margin py-4 max-w-container-max mx-auto h-20">
						<div className="flex items-center gap-8">
							<span className="font-display-pixel text-4xl uppercase text-on-surface">Flock</span>
							<div className="hidden lg:flex items-center gap-4">
								<a className="text-on-primary-fixed-variant font-bold border-2 border-on-background bg-primary-fixed shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] px-3 py-1 font-label-sm text-label-sm text-body-md" href="#">About</a>
								<a className="text-on-surface-variant font-medium px-3 py-1 hover:bg-surface-variant hover:translate-x-[-2px] hover:translate-y-[-2px] transition-all active:translate-x-[2px] active:translate-y-[2px] font-label-sm text-label-sm text-body-md" href="#">Events</a>
								<a className="text-on-surface-variant font-medium px-3 py-1 hover:bg-surface-variant hover:translate-x-[-2px] hover:translate-y-[-2px] transition-all active:translate-x-[2px] active:translate-y-[2px] font-label-sm text-label-sm text-body-md" href="#">Explore</a>
							</div>
						</div>
						<div className="flex items-center gap-4">
							<div className="hidden md:flex items-center border-2 border-on-background px-4 py-2 bg-surface">
								<span className="material-symbols-outlined mr-2">search</span>
								<input className="bg-transparent border-none focus:ring-0 font-label-sm text-label-sm p-0 w-32 lg:w-48" placeholder="Find your flock..." type="text" />
							</div>
							<a href="/dashboard" className="bg-primary-container text-on-primary-container border-2 border-on-background px-6 py-2 neo-shadow-sm font-bold active-push inline-block text-center">
								Join Flock
							</a>
						</div>
					</nav>
				</header>
				<main className="max-w-container-max mx-auto px-margin py-12 lg:py-24">
					{/* Hero Section */}
					<section className="text-center mb-24 relative">
						<div className="absolute -top-12 -left-12 opacity-10 pointer-events-none hidden lg:block">
							<span className="font-display-pixel text-[240px] leading-none uppercase">FLOCK</span>
						</div>
						<div className="relative z-10">
							<div className="inline-block bg-secondary-container border-2 border-on-background px-4 py-1 mb-8 neo-shadow-sm rotate-[-1deg]">
								<span className="font-label-sm text-label-sm text-on-secondary-container uppercase tracking-widest">Communities Reinvented</span>
							</div>
							<h1 className="font-headline-lg text-headline-lg lg:text-[100px] lg:leading-[100px] mb-8 max-w-4xl mx-auto uppercase lg:text-[120px] lg:leading-[110px]">
								Move Together, Find Your People.
							</h1>
							<p className="font-body-lg text-body-lg text-on-surface-variant max-w-2xl mx-auto mb-10">
								A community discovery platform that helps individuals find local groups, in-person events, and online communities where they can build genuine, long-lasting relationships.
							</p>
							<div className="flex justify-center items-center">
								<button className="bg-primary-container text-on-primary border-2 border-on-background px-10 py-5 text-xl font-bold neo-shadow active-push flex items-center gap-3">
									Find your community
									<span className="material-symbols-outlined">arrow_forward</span>
								</button>
							</div>
						</div>
					</section>
					{/* Bento Grid Section */}
					{/* Scroll-triggered Slideshow Effect Section */}
					<section className="grid grid-cols-2 lg:grid-cols-4 gap-gutter mb-32">
						<div className="border-2 border-on-background p-8 bg-surface text-center neo-shadow-sm">
							<span className="block font-display-pixel text-5xl mb-2 text-6xl">2.5K+</span>
							<p className="font-label-sm text-label-sm text-on-surface-variant uppercase font-bold text-body-md">Communities</p>
						</div>
						<div className="border-2 border-on-background p-8 bg-primary-fixed text-center neo-shadow-sm">
							<span className="block font-display-pixel text-5xl mb-2 text-6xl">180+</span>
							<p className="font-label-sm text-label-sm text-on-surface-variant uppercase font-bold text-body-md">Cities</p>
						</div>
						<div className="border-2 border-on-background p-8 bg-secondary-fixed text-center neo-shadow-sm">
							<span className="block font-display-pixel text-5xl mb-2 text-6xl">75+</span>
							<p className="font-label-sm text-label-sm text-on-surface-variant uppercase font-bold text-body-md">Interests</p>
						</div>
						<div className="border-2 border-on-background p-8 bg-tertiary-fixed text-center neo-shadow-sm">
							<span className="block font-display-pixel text-5xl mb-2 text-6xl">120K+</span>
							<p className="font-label-sm text-label-sm text-on-surface-variant uppercase font-bold text-body-md">Members</p>
						</div>
					</section>
					<div className="flex justify-end -mt-24 mb-24 px-2">
						<p className="font-label-sm text-[10px] text-on-surface-variant opacity-60 italic">
							*for representative purposes only
						</p>
					</div>
					<section className="mb-32 overflow-hidden border-4 border-on-background bg-surface-dim p-2 neo-shadow-lg">
						<div className="bg-background border-2 border-on-background py-16 px-8 relative">
							<div className="flex items-center justify-between mb-12 px-8">
								<h2 className="font-headline-lg text-headline-lg max-w-xl text-5xl">Find Your Vibe</h2>
								<div className="flex gap-4">
									<button onClick={() => scrollSlideshow(-1)} className="p-4 border-2 border-on-background bg-surface active-push hover:bg-surface-variant">
										<span className="material-symbols-outlined">chevron_left</span>
									</button>
									<button onClick={() => scrollSlideshow(1)} className="p-4 border-2 border-on-background bg-surface active-push hover:bg-surface-variant">
										<span className="material-symbols-outlined">chevron_right</span>
									</button>
								</div>
							</div>
							<div ref={slideshowRef} className="flex gap-gutter overflow-x-auto no-scrollbar pb-8 px-8 snap-x" id="slideshow">
								<div className="snap-center flex-none w-[300px] lg:w-[400px]">
									<div className="aspect-[4/5] border-2 border-on-background neo-shadow-sm relative mb-4">
										<img className="w-full h-full object-cover" data-alt="Artistic workshop setting with pottery wheels and clay, soft diffused natural light from large windows, minimalist aesthetic with pops of maroon and teal, clean black borders around every object." src="https://lh3.googleusercontent.com/aida-public/AB6AXuD9AEOoIPgzhqzhOOuhkVgX1YowRDbmQQUtHgw3TKJzmbcYUwNKt0u1ybxOlsXeJzorGRBL7dZ1wbbQbgP5_n3vUfwnaOIb0IM6HgLGZaG7DuQfdPips05FYr6bWARamc4XSbdoz1-17zQJShn6-BKPUXTQbV2Kg6qb1uj1p4QHQ4S07DP3CqLNXD_1zjgFG7FBAF44DLpMpz635YkxDiwx60YNRBZCZcNsbXZsdX8kdvuk9P6vEfx7gqqeM-915kzCnAyWOoy_eMOa" alt="" />
										<span className="absolute top-4 left-4 bg-tertiary-fixed border-2 border-on-background px-3 py-1 font-label-sm text-label-sm">Workshop</span>
									</div>
									<h4 className="font-headline-md text-headline-md text-3xl">Pottery &amp; Pints</h4>
									<p className="font-label-sm text-label-sm text-on-surface-variant uppercase text-body-md">Downtown Studio</p>
								</div>
								<div className="snap-center flex-none w-[300px] lg:w-[400px]">
									<div className="aspect-[4/5] border-2 border-on-background neo-shadow-sm relative mb-4">
										<img className="w-full h-full object-cover" data-alt="High-contrast close up of a retro synthesizer and mixing board, neon light accents in a dark studio room, neo-brutalist style product photography, sharp focus, technical atmosphere." src="https://lh3.googleusercontent.com/aida-public/AB6AXuCwCLvt55iLgXvTUzZ17edMOZxQFdV5gS1Fqj0CB4ozfE8zyiyIcNNRKt0FvzWZx4Oytb2lS0o6bXhAgrff9TumOrqNLNI3Iz03z491j47g_gUbN0aCZWCDEIIvLnEzu4h5sy0BVzBZ-5lWwlucdXxXEtjc341c6jWM0eOQIjg-yb1JRhLEGAiREH5_TnR_3nSMttqvb2XZUN4NSL7bF0WAk0wlVwit2WQ2jZEJyrkfs9-0bYoUOo9I90Wt13v9VoxcpSwNqAD9Bn86" alt="" />
										<span className="absolute top-4 left-4 bg-secondary-fixed border-2 border-on-background px-3 py-1 font-label-sm text-label-sm">Music Tech</span>
									</div>
									<h4 className="font-headline-md text-headline-md text-3xl">Synth Sundays</h4>
									<p className="font-label-sm text-label-sm text-on-surface-variant uppercase text-body-md">Community Center</p>
								</div>
								<div className="snap-center flex-none w-[300px] lg:w-[400px]">
									<div className="aspect-[4/5] border-2 border-on-background neo-shadow-sm relative mb-4">
										<img className="w-full h-full object-cover" data-alt="Outdoor public square filled with colorful market stalls and people browsing, bright midday sun, deep shadows, vibrant urban lifestyle photography, minimalist neo-brutalist framing." src="https://lh3.googleusercontent.com/aida-public/AB6AXuAKGf8AuqceZdLVY9gk-JJ4RTiYf5cHbItAkRfQSyNc_uAkCuSUm9uwzbHmZAjgIeyW-DpDuaT13EEynLq9yry58-lsUZrRlS0iIovmwwWu7tdoL8xarRYvpGXP5n6oTcDfGgCw5S7rqvD19ePfieIs94F01lEfnei38uDa3Yn-irZJfHSDVwkm4v4xkxN97BJMo4EJoDVhIrt4KOxr67OcwmTsGQW9ONb_g8wzHQyUHZKRh-Kh17qduCqHXY3Pmt3wkL2s1PR5kKxd" alt="" />
										<span className="absolute top-4 left-4 bg-primary-fixed border-2 border-on-background px-3 py-1 font-label-sm text-label-sm">Social</span>
									</div>
									<h4 className="font-headline-md text-headline-md text-3xl">Market Meetup</h4>
									<p className="font-label-sm text-label-sm text-on-surface-variant uppercase text-body-md">East Plaza</p>
								</div>
								<div className="snap-center flex-none w-[300px] lg:w-[400px]">
									<div className="aspect-[4/5] border-2 border-on-background neo-shadow-sm relative mb-4">
										<img className="w-full h-full object-cover" data-alt="Professional office setup with minimalist wooden desks and monitors, plants in geometric pots, high-key white lighting, sophisticated professional environment, clean black outlines." src="https://lh3.googleusercontent.com/aida-public/AB6AXuBH2RRVsXt0wiyhpClDCMiLPfOqJX5_8XO52DIFRlkWUXKxN1DOlPzLV89TrACZ6qprhxh0O9JEp0Gptrrz7rJyRsDY184qY5Yc_gJC0XkBhrms_wH9wQfQiBRmCyfQXFE4rsW9T1-4b4nu1EDoZpZBTjL63y-N5s83ET1sZ1H9Ue4CIsA5MpLyJPPQz7Ltt40s7JfS5kp8W0oHB5rj6VJk7mzmbydGPiBsjeUWPUXT_fQ4vMbX5WdL3Gzl8SehZt4LTqWBfOPVKZeN" alt="" />
										<span className="absolute top-4 left-4 bg-surface-variant border-2 border-on-background px-3 py-1 font-label-sm text-label-sm">Business</span>
									</div>
									<h4 className="font-headline-md text-headline-md text-3xl">Founder Circles</h4>
									<p className="font-label-sm text-label-sm text-on-surface-variant uppercase text-body-md">The Hub</p>
								</div>
							</div>
						</div>
					</section>
					{/* Call to Action Banner - Removed as requested */}
				</main>
				{/* Footer */}
				<footer className="bg-surface-container-highest border-t-4 border-on-background w-full">
					<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-gutter px-margin py-12 w-full max-w-container-max mx-auto">
						<div>
							<span className="font-headline-md text-headline-md font-black text-on-surface block mb-6 uppercase">FLOCK</span>
							<p className="font-body-md text-body-md text-on-surface-variant mb-6">Building the infrastructure for human connection in the digital age.</p>
							<div className="flex gap-4">
								<a className="w-10 h-10 bg-on-background flex items-center justify-center text-surface active-push" href="#">
									<span className="material-symbols-outlined">public</span>
								</a>
								<a className="w-10 h-10 bg-on-background flex items-center justify-center text-surface active-push" href="#">
									<span className="material-symbols-outlined">alternate_email</span>
								</a>
							</div>
						</div>
						<div>
							<h4 className="font-label-sm text-label-sm font-bold uppercase mb-6 text-on-surface text-body-md">Platform</h4>
							<ul className="space-y-3 font-label-sm text-label-sm text-body-md">
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Local Groups</a></li>
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Events</a></li>
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Online Hubs</a></li>
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Success Stories</a></li>
							</ul>
						</div>
						<div>
							<h4 className="font-label-sm text-label-sm font-bold uppercase mb-6 text-on-surface text-body-md">Community</h4>
							<ul className="space-y-3 font-label-sm text-label-sm text-body-md">
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Privacy Policy</a></li>
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Terms of Service</a></li>
								<li><a className="text-on-surface-variant hover:text-primary transition-colors font-bold underline text-primary" href="#">Community Guidelines</a></li>
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Safety Center</a></li>
							</ul>
						</div>
						<div>
							<h4 className="font-label-sm text-label-sm font-bold uppercase mb-6 text-on-surface text-body-md">Resources</h4>
							<ul className="space-y-3 font-label-sm text-label-sm text-body-md">
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Help Center</a></li>
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Contact Us</a></li>
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Twitter</a></li>
								<li><a className="text-on-surface-variant hover:text-primary transition-colors" href="#">Discord</a></li>
							</ul>
						</div>
					</div>
					<div className="px-margin py-8 border-t-2 border-on-background max-w-container-max mx-auto flex flex-col md:flex-row justify-between items-center gap-4">
						<p className="font-label-sm text-label-sm text-on-surface-variant">© 2024 Flock Community Discovery. Built for Connection.</p>
						<div className="flex gap-6 font-label-sm text-label-sm">
						</div>
					</div>
				</footer>
			</div>
		</>
	);
}
