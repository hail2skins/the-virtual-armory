package payment_test_utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/models"
)

// canSubscribeToTier checks if a user can subscribe to a specific tier based on their current subscription
func canSubscribeToTier(user *models.User, tier string) bool {
	if user == nil {
		return true // Guest users can subscribe to any tier
	}

	// Users can always upgrade to a higher tier
	switch user.SubscriptionTier {
	case "free", "":
		return true // Free users can subscribe to any tier
	case "monthly":
		// Monthly users can upgrade to yearly, lifetime, or premium_lifetime
		return tier != "monthly"
	case "yearly":
		// Yearly users can upgrade to lifetime or premium_lifetime
		return tier != "monthly" && tier != "yearly"
	case "lifetime":
		// Lifetime users can only upgrade to premium_lifetime
		return tier == "premium_lifetime"
	case "premium_lifetime":
		// Premium lifetime users cannot subscribe to any other tier
		return false
	default:
		return true
	}
}

// MockPricingPage renders a mock pricing page for tests
func MockPricingPage(c *gin.Context, user *models.User) {
	// Determine if the user has a subscription
	var currentPlanHTML string
	if user != nil {
		if user.SubscriptionTier == "free" || user.SubscriptionTier == "" {
			currentPlanHTML = `<div class="mt-8"><div class="bg-gray-200 border border-gray-300 text-gray-800 font-semibold py-2 px-4 rounded text-center">Current Plan</div></div>`
		} else if user.SubscriptionTier == "monthly" {
			currentPlanHTML = `<div class="mt-8"><div class="bg-indigo-200 border border-indigo-300 text-indigo-800 font-semibold py-2 px-4 rounded text-center">Current Plan</div></div>`
		} else if user.SubscriptionTier == "yearly" {
			currentPlanHTML = `<div class="mt-8"><div class="bg-green-200 border border-green-300 text-green-800 font-semibold py-2 px-4 rounded text-center">Current Plan</div></div>`
		} else if user.SubscriptionTier == "lifetime" {
			currentPlanHTML = `<div class="mt-8"><div class="bg-purple-200 border border-purple-300 text-purple-800 font-semibold py-2 px-4 rounded text-center">Current Plan</div></div>`
		} else if user.SubscriptionTier == "premium_lifetime" {
			currentPlanHTML = `<div class="mt-8"><div class="bg-yellow-200 border border-yellow-300 text-yellow-800 font-semibold py-2 px-4 rounded text-center">Current Plan</div></div>`
		}
	}

	// HTML for disabled subscription button
	disabledButtonHTML := `<div class="mt-8"><div class="block w-full bg-gray-400 text-white font-semibold py-2 px-4 rounded cursor-not-allowed text-center">Already subscribed to a higher tier</div></div>`

	// Generate the free plan button/current plan indicator
	freePlanHTML := `<div class="mt-8"><div class="text-gray-500 font-medium py-2 px-4 text-center">Default Free Plan</div></div>`
	if user != nil && (user.SubscriptionTier == "free" || user.SubscriptionTier == "") {
		freePlanHTML = currentPlanHTML
	}

	// Generate the monthly plan button/current plan indicator
	monthlyPlanHTML := `<div class="mt-8"><a href="/checkout?tier=monthly" class="block w-full bg-indigo-600 text-white font-semibold py-2 px-4 rounded hover:bg-indigo-700 transition duration-200 text-center">Subscribe Monthly</a></div>`
	if user != nil {
		if user.SubscriptionTier == "monthly" {
			monthlyPlanHTML = currentPlanHTML
		} else if !canSubscribeToTier(user, "monthly") {
			monthlyPlanHTML = disabledButtonHTML
		}
	}

	// Generate the yearly plan button/current plan indicator
	yearlyPlanHTML := `<div class="mt-8"><a href="/checkout?tier=yearly" class="block w-full bg-green-600 text-white font-semibold py-2 px-4 rounded hover:bg-green-700 transition duration-200 text-center">Subscribe Yearly</a></div>`
	if user != nil {
		if user.SubscriptionTier == "yearly" {
			yearlyPlanHTML = currentPlanHTML
		} else if !canSubscribeToTier(user, "yearly") {
			yearlyPlanHTML = disabledButtonHTML
		}
	}

	// Generate the lifetime plan button/current plan indicator
	lifetimePlanHTML := `<div class="mt-8"><a href="/checkout?tier=lifetime" class="block w-full bg-purple-600 text-white font-semibold py-2 px-4 rounded hover:bg-purple-700 transition duration-200 text-center">Buy Lifetime</a></div>`
	if user != nil {
		if user.SubscriptionTier == "lifetime" {
			lifetimePlanHTML = currentPlanHTML
		} else if !canSubscribeToTier(user, "lifetime") {
			lifetimePlanHTML = disabledButtonHTML
		}
	}

	// Generate the premium lifetime plan button/current plan indicator
	premiumLifetimePlanHTML := `<div class="mt-8"><a href="/checkout?tier=premium_lifetime" class="block w-full bg-yellow-600 text-white font-semibold py-2 px-4 rounded hover:bg-yellow-700 transition duration-200 text-center">Buy Premium Lifetime</a></div>`
	if user != nil && user.SubscriptionTier == "premium_lifetime" {
		premiumLifetimePlanHTML = `<div class="mt-8"><div class="bg-yellow-200 border border-yellow-300 text-yellow-800 font-semibold py-2 px-4 rounded text-center">Current Plan</div></div>`
	}

	content := fmt.Sprintf(`
		<div class="bg-white py-12">
			<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<!-- Header -->
				<div class="text-center mb-12">
					<h2 class="text-3xl font-extrabold text-gray-900 sm:text-4xl">
						Simple, transparent pricing
					</h2>
					<p class="mt-4 text-lg text-gray-600">
						Choose Your Plan
					</p>
				</div>
				
				<!-- Pricing Cards -->
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
					<!-- Free Plan -->
					<div class="border border-gray-200 rounded-lg shadow-sm p-6 bg-white">
						<h2 class="text-2xl font-semibold text-gray-900">Free</h2>
						<p class="mt-4 text-sm text-gray-500">Basic access</p>
						<p class="mt-8"><span class="text-4xl font-extrabold text-gray-900">$0</span> <span class="text-base font-medium text-gray-500">/forever</span></p>
						<ul class="mt-6 space-y-3">
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Store up to 2 guns</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Store up to 4 ammunition</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Limited range days</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">No maintenance records</span></li>
						</ul>
						%s
					</div>
					
					<!-- Monthly Plan -->
					<div class="border border-gray-200 rounded-lg shadow-sm p-6 bg-white">
						<h2 class="text-2xl font-semibold text-gray-900">Liking It</h2>
						<p class="mt-4 text-sm text-gray-500">Flexible option</p>
						<p class="mt-8"><span class="text-4xl font-extrabold text-gray-900">$5</span> <span class="text-base font-medium text-gray-500">/mo</span></p>
						<ul class="mt-6 space-y-3">
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Unlimited guns/ammo</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Unlimited range days</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Unlimited maintenance records</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Cancel anytime</span></li>
						</ul>
						%s
					</div>
					
					<!-- Yearly Plan -->
					<div class="border border-gray-200 rounded-lg shadow-sm p-6 bg-white">
						<h2 class="text-2xl font-semibold text-gray-900">Loving It</h2>
						<p class="mt-4 text-sm text-gray-500">Best value</p>
						<p class="mt-8"><span class="text-4xl font-extrabold text-gray-900">$30</span> <span class="text-base font-medium text-gray-500">/year</span></p>
						<ul class="mt-6 space-y-3">
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Unlimited guns/ammo</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Unlimited range days</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Unlimited maintenance records</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Save $30 vs monthly</span></li>
						</ul>
						%s
					</div>
					
					<!-- Lifetime Plan -->
					<div class="border border-gray-200 rounded-lg shadow-sm p-6 bg-white">
						<h2 class="text-2xl font-semibold text-gray-900">Supporter</h2>
						<p class="mt-4 text-sm text-gray-500">Lifetime access</p>
						<p class="mt-8"><span class="text-4xl font-extrabold text-gray-900">$100</span> <span class="text-base font-medium text-gray-500">/lifetime</span></p>
						<ul class="mt-6 space-y-3">
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Unlimited guns/ammo</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Unlimited range days</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Unlimited maintenance records</span></li>
							<li class="flex items-start"><span class="text-green-500 flex-shrink-0 mr-2">✓</span> <span class="text-sm text-gray-500">Never pay again</span></li>
						</ul>
						%s
					</div>
				</div>

				<!-- Premium Lifetime Plan -->
				<div class="mt-8">
					<div class="bg-gradient-to-r from-purple-600 to-indigo-600 rounded-lg shadow-lg overflow-hidden">
						<div class="p-8">
							<h2 class="text-3xl font-extrabold text-white">
								Big Baller
								<span class="block text-lg font-medium mt-1">You shouldn't have, but thanks.</span>
							</h2>
							<p class="mt-4 text-lg leading-6 text-white">
								For our biggest supporters who want to help us grow.
							</p>
							<ul class="mt-8 space-y-4">
								<li class="flex items-start">
									<span class="text-white flex-shrink-0 mr-2">✓</span>
									<span class="text-base font-medium text-white">Everything the site has.</span>
								</li>
								<li class="flex items-start">
									<span class="text-white flex-shrink-0 mr-2">✓</span>
									<span class="text-base font-medium text-white">Christmas cards. Seriously, send your address and they are yours.</span>
								</li>
								<li class="flex items-start">
									<span class="text-white flex-shrink-0 mr-2">✓</span>
									<span class="text-base font-medium text-white">If it grows and makers provide goodies, they go to you first.</span>
								</li>
							</ul>
							%s
						</div>
					</div>
				</div>
			</div>
		</div>
	`, freePlanHTML, monthlyPlanHTML, yearlyPlanHTML, lifetimePlanHTML, premiumLifetimePlanHTML)

	// Use a simple HTML renderer for tests
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteHeader(200)
	c.Writer.WriteString("<html>" + content + "</html>")
}
