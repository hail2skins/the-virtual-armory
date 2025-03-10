package payment_test_utils

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/models"
)

// MockPaymentHistory renders a mock payment history page for tests
func MockPaymentHistory(c *gin.Context, user *models.User) {
	// Create a mock payment history response
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteHeader(200)

	// Get the subscription tier name
	var tierName string
	switch user.SubscriptionTier {
	case "free":
		tierName = "Free Tier"
	case "monthly":
		tierName = "Liking It"
	case "yearly":
		tierName = "Loving It"
	case "lifetime":
		tierName = "Supporter"
	case "premium_lifetime":
		tierName = "Big Baller"
	default:
		tierName = "Unknown Tier"
	}

	// Get the subscription benefits
	var benefits string
	switch user.SubscriptionTier {
	case "free":
		benefits = `
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Store up to 2 guns</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Store up to 4 ammunition*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Limited range days*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>No maintenance records*</span>
		</li>
		`
	case "monthly":
		benefits = `
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Unlimited guns/ammo*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Unlimited range days*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Unlimited maintenance records*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Cancel anytime</span>
		</li>
		`
	case "yearly":
		benefits = `
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Unlimited guns/ammo*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Unlimited range days*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Unlimited maintenance records*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Cancel anytime</span>
		</li>
		`
	case "lifetime":
		benefits = `
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Unlimited guns/ammo*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Unlimited range days*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Unlimited maintenance records*</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>First to access new features*</span>
		</li>
		`
	case "premium_lifetime":
		benefits = `
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Everything the site has</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Christmas cards. Seriously, send your address and they are yours.</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>If it grows and makers provide goodies, they go to you first.</span>
		</li>
		<li class="flex items-start text-gray-700">
			<span class="text-green-500 mr-2">✓</span>
			<span>Premium support and early access to new features</span>
		</li>
		`
	}

	// Create the HTML
	html := `
	<div class="bg-white rounded-xl shadow-lg overflow-hidden mb-10">
		<div class="bg-blue-600 p-6 text-white">
			<h2 class="text-xl font-semibold">Current Subscription</h2>
		</div>
		<div class="p-6">
			<div class="flex flex-col md:flex-row justify-between items-start md:items-center mb-6">
				<div>
					<h3 class="text-lg font-medium text-gray-900">
						` + tierName + `
					</h3>
				</div>
			</div>
			
			<div class="border-t border-gray-200 pt-6">
				<h4 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-3">Subscription Benefits</h4>
				<ul class="space-y-3">
					` + benefits + `
				</ul>
				<div class="mt-4 text-sm text-gray-500">* = When available</div>
			</div>
		</div>
	</div>
	`

	c.Writer.WriteString(html)
}
