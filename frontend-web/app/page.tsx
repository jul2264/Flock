"use client";

import Link from "next/link";
import { useAuth, SignInButton } from "@clerk/nextjs";

export default function Home() {
	const { isSignedIn } = useAuth();

	return (
		<div className="relative min-h-screen flex flex-col justify-between overflow-hidden bg-zinc-950 text-zinc-50 font-sans selection:bg-violet-500/35 selection:text-white">
			{/* Radiant Background Effects */}
			<div className="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] rounded-full bg-violet-600/10 blur-[120px] pointer-events-none" />
			<div className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] rounded-full bg-indigo-600/10 blur-[120px] pointer-events-none" />
			<div className="absolute inset-0 bg-[linear-gradient(to_right,#1f1f23_1px,transparent_1px),linear-gradient(to_bottom,#1f1f23_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_50%,#000_70%,transparent_100%)] opacity-35" />

			{/* Navbar */}
			<header className="relative z-10 w-full max-w-7xl mx-auto px-6 py-6 flex items-center justify-between border-b border-zinc-900 bg-zinc-950/20 backdrop-blur-md">
				<div className="flex items-center gap-2">
					<div className="h-8 w-8 rounded-lg bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center font-bold text-white shadow-lg shadow-violet-500/20">
						F
					</div>
					<span className="text-xl font-bold tracking-tight bg-gradient-to-r from-white via-zinc-200 to-zinc-400 bg-clip-text text-transparent">
						Flock
					</span>
				</div>
				<nav className="flex items-center gap-4">
					<a
						href="https://github.com/jul2264/Flock"
						target="_blank"
						rel="noopener noreferrer"
						className="text-sm font-medium text-zinc-400 hover:text-white transition-colors"
					>
						GitHub
					</a>
					{!isSignedIn ? (
						<SignInButton mode="modal">
							<button className="cursor-pointer text-sm font-medium text-zinc-300 hover:text-white transition-colors bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 rounded-full px-4 py-2">
								Sign In
							</button>
						</SignInButton>
					) : (
						<Link
							href="/dashboard"
							className="text-sm font-medium text-white transition-all bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 rounded-full px-4 py-2 shadow-lg shadow-violet-600/25"
						>
							Console
						</Link>
					)}
				</nav>
			</header>

			{/* Hero Section */}
			<main className="relative z-10 flex-1 max-w-7xl mx-auto px-6 flex flex-col justify-center items-center text-center">
				<div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-violet-500/20 bg-violet-500/5 text-violet-400 text-xs font-semibold uppercase tracking-wider mb-8 animate-pulse">
					<span>✨</span> Phase 6: Administration & Moderation Active
				</div>
				<h1 className="max-w-4xl text-5xl md:text-7xl font-extrabold tracking-tight bg-gradient-to-b from-white via-zinc-100 to-zinc-400 bg-clip-text text-transparent leading-[1.1] mb-6">
					Move Together, <br />
					<span className="bg-gradient-to-r from-violet-400 via-fuchsia-400 to-indigo-400 bg-clip-text text-transparent">
						Find Your People.
					</span>
				</h1>
				<p className="max-w-2xl text-lg text-zinc-400 leading-relaxed mb-12">
					Welcome to the official Flock Management Console. Oversee community activities, moderate events, promote organizer roles, and analyze platform telemetry.
				</p>

				<div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
					{!isSignedIn ? (
						<SignInButton mode="modal">
							<button className="cursor-pointer flex h-14 items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 px-8 text-base font-semibold text-white shadow-xl shadow-violet-600/25 hover:scale-[1.02] active:scale-[0.98] transition-all hover:shadow-violet-600/35">
								Enter Organizer Console
							</button>
						</SignInButton>
					) : (
						<Link
							href="/dashboard"
							className="flex h-14 items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 px-8 text-base font-semibold text-white shadow-xl shadow-violet-600/25 hover:scale-[1.02] active:scale-[0.98] transition-all hover:shadow-violet-600/35"
						>
							Enter Dashboard
						</Link>
					)}
					<Link
						href="/dashboard"
						className="flex h-14 items-center justify-center rounded-xl border border-zinc-800 bg-zinc-950/40 px-8 text-base font-semibold text-zinc-300 hover:text-white transition-all hover:bg-zinc-900/60 hover:border-zinc-700"
					>
						Platform Metrics
					</Link>
				</div>

				{/* Feature Cards Grid */}
				<div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-5xl w-full mt-24">
					<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/40 backdrop-blur-md text-left hover:border-violet-500/20 transition-all duration-300">
						<div className="h-10 w-10 rounded-xl bg-violet-600/10 border border-violet-500/25 flex items-center justify-center text-violet-400 font-bold mb-4">
							📊
						</div>
						<h3 className="text-lg font-semibold mb-2">Live Analytics</h3>
						<p className="text-sm text-zinc-400">
							Track real-time active users, RSVP conversion rates, popular interests, and community engagement.
						</p>
					</div>

					<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/40 backdrop-blur-md text-left hover:border-fuchsia-500/20 transition-all duration-300">
						<div className="h-10 w-10 rounded-xl bg-fuchsia-600/10 border border-fuchsia-500/25 flex items-center justify-center text-fuchsia-400 font-bold mb-4">
							🛡️
						</div>
						<h3 className="text-lg font-semibold mb-2">Role Moderation</h3>
						<p className="text-sm text-zinc-400">
							Promote regular users to organizers, manage administrative access levels, and ensure trust.
						</p>
					</div>

					<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/40 backdrop-blur-md text-left hover:border-indigo-500/20 transition-all duration-300">
						<div className="h-10 w-10 rounded-xl bg-indigo-600/10 border border-indigo-500/25 flex items-center justify-center text-indigo-400 font-bold mb-4">
							⚡
						</div>
						<h3 className="text-lg font-semibold mb-2">Realtime Sync</h3>
						<p className="text-sm text-zinc-400">
							Moderate reported events and deactivate spam instantly, fully synchronized with Meilisearch.
						</p>
					</div>
				</div>
			</main>

			{/* Footer */}
			<footer className="relative z-10 w-full max-w-7xl mx-auto px-6 py-8 flex flex-col sm:flex-row items-center justify-between gap-4 border-t border-zinc-900 mt-16 text-sm text-zinc-500">
				<p>© {new Date().getFullYear()} Flock Inc. All rights reserved.</p>
				<p>Hyperlocal Social Discovery Platform</p>
			</footer>
		</div>
	);
}
