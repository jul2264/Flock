"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuth } from "@clerk/nextjs";
import { api, User } from "../../lib/api";
import { 
	Calendar, 
	Users, 
	ShieldAlert, 
	ArrowRight, 
	Info,
	Lock,
	CheckCircle
} from "lucide-react";

export default function DashboardPage() {
	const { getToken, isSignedIn } = useAuth();
	const [user, setUser] = useState<User | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		async function fetchProfile() {
			try {
				const token = await getToken();
				if (token) {
					const profile = await api.getMe(token);
					setUser(profile);
				}
			} catch (err) {
				console.error("Failed to load user profile in dashboard page:", err);
			} finally {
				setLoading(false);
			}
		}

		if (isSignedIn) {
			fetchProfile();
		}
	}, [isSignedIn, getToken]);

	if (loading) {
		return (
			<div className="flex flex-col items-center justify-center min-h-[50vh] gap-4">
				<div className="h-8 w-8 border-4 border-violet-600 border-t-transparent rounded-full animate-spin" />
				<p className="text-zinc-500 text-sm">Loading console summary...</p>
			</div>
		);
	}

	if (!user) return null;

	const isOrganizer = user.role === "organizer" || user.role === "admin";
	const isAdmin = user.role === "admin";

	return (
		<div className="space-y-8">
			{/* Top Hero Banner */}
			<div className="relative overflow-hidden p-8 md:p-10 rounded-3xl border border-violet-500/10 bg-gradient-to-tr from-zinc-950 via-zinc-900 to-violet-950/20">
				<div className="absolute top-[-20%] right-[-10%] w-[350px] h-[350px] rounded-full bg-violet-600/10 blur-[100px] pointer-events-none" />
				<div className="relative z-10 max-w-2xl text-left space-y-4">
					<span className="text-xs font-semibold uppercase tracking-wider text-violet-400 bg-violet-500/10 border border-violet-500/25 px-3 py-1 rounded-full">
						Flock System Console
					</span>
					<h1 className="text-3xl md:text-5xl font-extrabold tracking-tight text-white leading-tight">
						Welcome Back, <br />
						<span className="bg-gradient-to-r from-violet-400 to-indigo-400 bg-clip-text text-transparent">
							{user.full_name || "Console User"}
						</span>
					</h1>
					<p className="text-zinc-400 text-sm md:text-base leading-relaxed">
						This console lets you create events, manage communities, moderate RSVPs, and manage user access permissions across the Flock hyperlocal platform.
					</p>
				</div>
			</div>

			{/* Role Card / Status Info */}
			<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
				{/* Identity Panel */}
				<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/30 backdrop-blur-md space-y-4 text-left">
					<h2 className="text-lg font-bold text-zinc-200">Your Identity</h2>
					<div className="space-y-3">
						<div className="flex justify-between items-center py-2 border-b border-zinc-900">
							<span className="text-sm text-zinc-500">Email Address</span>
							<span className="text-sm text-zinc-300 font-medium">{user.email}</span>
						</div>
						<div className="flex justify-between items-center py-2 border-b border-zinc-900">
							<span className="text-sm text-zinc-500">Current Role</span>
							<span className="text-sm text-violet-400 font-semibold uppercase tracking-wider">
								{user.role}
							</span>
						</div>
						<div className="flex justify-between items-center py-2 border-b border-zinc-900">
							<span className="text-sm text-zinc-500">User ID</span>
							<span className="text-xs text-zinc-500 font-mono select-all truncate max-w-[150px]">
								{user.id}
							</span>
						</div>
						<div className="flex justify-between items-center py-2">
							<span className="text-sm text-zinc-500">Status</span>
							<span className="text-xs font-semibold text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded-full flex items-center gap-1">
								<CheckCircle className="h-3 w-3" /> Active
							</span>
						</div>
					</div>
				</div>

				{/* Help & Promotion Requests */}
				<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/30 backdrop-blur-md flex flex-col justify-between text-left">
					<div className="space-y-3">
						<h2 className="text-lg font-bold text-zinc-200">Role Guidelines</h2>
						<p className="text-sm text-zinc-400 leading-relaxed">
							Regular **Users** can view meetups on the mobile client. <br />
							**Organizers** can create and update events or publish community announcements. <br />
							**Admins** moderate the platform, configure index settings, and promote user roles.
						</p>
					</div>

					{!isOrganizer && (
						<div className="mt-6 p-4 rounded-xl border border-amber-500/10 bg-amber-500/5 text-xs text-amber-400 flex items-start gap-3">
							<Info className="h-4 w-4 flex-shrink-0 mt-0.5" />
							<span>
								You currently have a standard <strong>User</strong> role. If you want to build communities or organize events, please contact a platform Administrator to request an upgrade to <strong>Organizer</strong>.
							</span>
						</div>
					)}
				</div>
			</div>

			{/* Quick Action Navigation Grid */}
			<div className="space-y-4">
				<h2 className="text-xl font-bold text-zinc-100 text-left">Quick Navigation</h2>
				<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
					{/* Organizer Dashboard Link */}
					<div className={`p-6 rounded-2xl border transition-all ${
						isOrganizer 
							? "border-zinc-900 bg-zinc-950/30 hover:border-violet-500/20" 
							: "border-zinc-900/50 bg-zinc-950/10 opacity-60"
					} flex flex-col justify-between h-48 text-left group`}>
						<div>
							<div className="h-10 w-10 rounded-xl bg-violet-600/10 border border-violet-500/20 flex items-center justify-center text-violet-400 mb-4">
								<Calendar className="h-5 w-5" />
							</div>
							<h3 className="text-lg font-bold text-zinc-200 group-hover:text-violet-400 transition-colors">
								Organizer Portal
							</h3>
							<p className="text-sm text-zinc-400 mt-1">
								Manage events, review RSVPs, and post community announcements.
							</p>
						</div>
						{isOrganizer ? (
							<Link href="/dashboard/organizer" className="flex items-center gap-1.5 text-sm font-semibold text-violet-400 hover:text-violet-300 mt-4 self-start">
								Go to Portal <ArrowRight className="h-4 w-4 group-hover:translate-x-1 transition-transform" />
							</Link>
						) : (
							<div className="flex items-center gap-1.5 text-xs text-zinc-600 mt-4">
								<Lock className="h-3.5 w-3.5" /> Requires Organizer Privilege
							</div>
						)}
					</div>

					{/* Admin Console Link */}
					<div className={`p-6 rounded-2xl border transition-all ${
						isAdmin 
							? "border-zinc-900 bg-zinc-950/30 hover:border-red-500/20" 
							: "border-zinc-900/50 bg-zinc-950/10 opacity-60"
					} flex flex-col justify-between h-48 text-left group`}>
						<div>
							<div className="h-10 w-10 rounded-xl bg-red-600/10 border border-red-500/20 flex items-center justify-center text-red-400 mb-4">
								<ShieldAlert className="h-5 w-5" />
							</div>
							<h3 className="text-lg font-bold text-zinc-200 group-hover:text-red-400 transition-colors">
								Admin Console
							</h3>
							<p className="text-sm text-zinc-400 mt-1">
								Analyze platform stats, manage roles, and moderate events.
							</p>
						</div>
						{isAdmin ? (
							<Link href="/dashboard/admin" className="flex items-center gap-1.5 text-sm font-semibold text-red-400 hover:text-red-300 mt-4 self-start">
								Go to Console <ArrowRight className="h-4 w-4 group-hover:translate-x-1 transition-transform" />
							</Link>
						) : (
							<div className="flex items-center gap-1.5 text-xs text-zinc-600 mt-4">
								<Lock className="h-3.5 w-3.5" /> Requires Admin Privilege
							</div>
						)}
					</div>
				</div>
			</div>
		</div>
	);
}
