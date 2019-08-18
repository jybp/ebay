package ebay

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// BrowseService handles communication with the Browse API.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/browse/overview.html
type BrowseService service

// Valid values for the "buyingOptions" item field.
const (
	BrowseBuyingOptionAuction    = "AUCTION"
	BrowseBuyingOptionFixedPrice = "FIXED_PRICE"
)

// OptBrowseContextualLocation adds the header containing contextualLocation.
// It is strongly recommended that you use it when submitting Browse API methods.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/static/api-browse.html#Headers
func OptBrowseContextualLocation(country, zip string) func(*http.Request) {
	return func(req *http.Request) {
		const headerEndUserCtx = "X-EBAY-C-ENDUSERCTX"
		v := req.Header.Get(headerEndUserCtx)
		if len(v) > 0 {
			v += ","
		}
		v += "contextualLocation=" + url.QueryEscape(fmt.Sprintf("country=%s,zip=%s", country, zip))
		req.Header.Set(headerEndUserCtx, v)
	}
}

// LegacyItem represents the legacy representation of an eBay item.
type LegacyItem struct {
	ItemID             string `json:"itemId"`
	SellerItemRevision string `json:"sellerItemRevision"`
	Title              string `json:"title"`
	ShortDescription   string `json:"shortDescription"`
	Price              struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"price"`
	CategoryPath string `json:"categoryPath"`
	Condition    string `json:"condition"`
	ConditionID  string `json:"conditionId"`
	ItemLocation struct {
		City            string `json:"city"`
		StateOrProvince string `json:"stateOrProvince"`
		PostalCode      string `json:"postalCode"`
		Country         string `json:"country"`
	} `json:"itemLocation"`
	Image struct {
		ImageURL string `json:"imageUrl"`
	} `json:"image"`
	AdditionalImages []struct {
		ImageURL string `json:"imageUrl"`
	} `json:"additionalImages"`
	Brand       string    `json:"brand"`
	ItemEndDate time.Time `json:"itemEndDate"`
	Seller      struct {
		Username           string `json:"username"`
		FeedbackPercentage string `json:"feedbackPercentage"`
		FeedbackScore      int    `json:"feedbackScore"`
	} `json:"seller"`
	Gtin                    string `json:"gtin"`
	EstimatedAvailabilities []struct {
		DeliveryOptions             []string `json:"deliveryOptions"`
		EstimatedAvailabilityStatus string   `json:"estimatedAvailabilityStatus"`
		EstimatedAvailableQuantity  int      `json:"estimatedAvailableQuantity"`
		EstimatedSoldQuantity       int      `json:"estimatedSoldQuantity"`
	} `json:"estimatedAvailabilities"`
	ShippingOptions []struct {
		ShippingServiceCode string `json:"shippingServiceCode"`
		TrademarkSymbol     string `json:"trademarkSymbol"`
		ShippingCarrierCode string `json:"shippingCarrierCode"`
		Type                string `json:"type"`
		ShippingCost        struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"shippingCost"`
		QuantityUsedForEstimate       int       `json:"quantityUsedForEstimate"`
		MinEstimatedDeliveryDate      time.Time `json:"minEstimatedDeliveryDate"`
		MaxEstimatedDeliveryDate      time.Time `json:"maxEstimatedDeliveryDate"`
		ShipToLocationUsedForEstimate struct {
			PostalCode string `json:"postalCode"`
			Country    string `json:"country"`
		} `json:"shipToLocationUsedForEstimate"`
		AdditionalShippingCostPerUnit struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"additionalShippingCostPerUnit"`
		ShippingCostType string `json:"shippingCostType"`
	} `json:"shippingOptions"`
	ShipToLocations struct {
		RegionIncluded []struct {
			RegionName string `json:"regionName"`
			RegionType string `json:"regionType"`
		} `json:"regionIncluded"`
		RegionExcluded []struct {
			RegionName string `json:"regionName"`
			RegionType string `json:"regionType"`
		} `json:"regionExcluded"`
	} `json:"shipToLocations"`
	ReturnTerms struct {
		ReturnsAccepted         bool   `json:"returnsAccepted"`
		RefundMethod            string `json:"refundMethod"`
		ReturnMethod            string `json:"returnMethod"`
		ReturnShippingCostPayer string `json:"returnShippingCostPayer"`
		ReturnPeriod            struct {
			Value int    `json:"value"`
			Unit  string `json:"unit"`
		} `json:"returnPeriod"`
		RestockingFeePercentage string `json:"restockingFeePercentage"`
	} `json:"returnTerms"`
	Taxes []struct {
		TaxJurisdiction struct {
			Region struct {
				RegionName string `json:"regionName"`
				RegionType string `json:"regionType"`
			} `json:"region"`
			TaxJurisdictionID string `json:"taxJurisdictionId"`
		} `json:"taxJurisdiction"`
		TaxType                  string `json:"taxType"`
		TaxPercentage            string `json:"taxPercentage"`
		ShippingAndHandlingTaxed bool   `json:"shippingAndHandlingTaxed"`
		IncludedInPrice          bool   `json:"includedInPrice"`
	} `json:"taxes"`
	LocalizedAspects []struct {
		Type  string `json:"type"`
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"localizedAspects"`
	PrimaryProductReviewRating struct {
		ReviewCount      int    `json:"reviewCount"`
		AverageRating    string `json:"averageRating"`
		RatingHistograms []struct {
			Rating string `json:"rating"`
			Count  int    `json:"count"`
		} `json:"ratingHistograms"`
	} `json:"primaryProductReviewRating"`
	TopRatedBuyingExperience bool     `json:"topRatedBuyingExperience"`
	BuyingOptions            []string `json:"buyingOptions"`
	ItemAffiliateWebURL      string   `json:"itemAffiliateWebUrl"`
	ItemWebURL               string   `json:"itemWebUrl"`
	Description              string   `json:"description"`
	EnabledForGuestCheckout  bool     `json:"enabledForGuestCheckout"`
	AdultOnly                bool     `json:"adultOnly"`
	CategoryID               string   `json:"categoryId"`
}

// GetItemByLegacyID retrieves an item by legacy ID.
// The itemID will be available in the "itemId" field:
// https://developer.ebay.com/api-docs/buy/static/api-browse.html#Legacy
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/browse/resources/item/methods/getItemByLegacyId
func (s *BrowseService) GetItemByLegacyID(ctx context.Context, itemLegacyID string, opts ...Opt) (CompactItem, error) {
	u := fmt.Sprintf("buy/browse/v1/item/get_item_by_legacy_id?legacy_item_id=%s", itemLegacyID)
	req, err := s.client.NewRequest(http.MethodGet, u, nil, opts...)
	if err != nil {
		return CompactItem{}, err
	}
	var it CompactItem
	return it, s.client.Do(ctx, req, &it)
}

// CompactItem represents the "COMPACT" version of an eBay item.
type CompactItem struct {
	ItemID             string `json:"itemId"`
	SellerItemRevision string `json:"sellerItemRevision"`
	Price              struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"price"`
	EstimatedAvailabilities []struct {
		AvailabilityThresholdType   string `json:"availabilityThresholdType"`
		AvailabilityThreshold       int    `json:"availabilityThreshold"`
		EstimatedAvailabilityStatus string `json:"estimatedAvailabilityStatus"`
		EstimatedSoldQuantity       int    `json:"estimatedSoldQuantity"`
	} `json:"estimatedAvailabilities"`
	TopRatedBuyingExperience bool `json:"topRatedBuyingExperience"`
}

// GetCompactItem retrieves the compact version of a specific item.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/browse/resources/item/methods/getItem
func (s *BrowseService) GetCompactItem(ctx context.Context, itemID string, opts ...Opt) (CompactItem, error) {
	u := fmt.Sprintf("buy/browse/v1/item/%s?fieldgroups=COMPACT", itemID)
	req, err := s.client.NewRequest(http.MethodGet, u, nil, opts...)
	if err != nil {
		return CompactItem{}, err
	}
	var it CompactItem
	return it, s.client.Do(ctx, req, &it)
}

// Item represents an eBay item.
type Item struct {
	ItemID             string `json:"itemId"`
	SellerItemRevision string `json:"sellerItemRevision"`
	Title              string `json:"title"`
	Subtitle           string `json:"subtitle"`
	ShortDescription   string `json:"shortDescription"`
	Price              struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"price"`
	CategoryPath string `json:"categoryPath"`
	Condition    string `json:"condition"`
	ConditionID  string `json:"conditionId"`
	ItemLocation struct {
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"itemLocation"`
	Image struct {
		ImageURL string `json:"imageUrl"`
	} `json:"image"`
	AdditionalImages []struct {
		ImageURL string `json:"imageUrl"`
	} `json:"additionalImages"`
	MarketingPrice struct {
		OriginalPrice struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"originalPrice"`
		DiscountPercentage string `json:"discountPercentage"`
		DiscountAmount     struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"discountAmount"`
	} `json:"marketingPrice"`
	Color  string `json:"color"`
	Brand  string `json:"brand"`
	Seller struct {
		Username           string `json:"username"`
		FeedbackPercentage string `json:"feedbackPercentage"`
		FeedbackScore      int    `json:"feedbackScore"`
	} `json:"seller"`
	Gtin                    string `json:"gtin"`
	Mpn                     string `json:"mpn"`
	Epid                    string `json:"epid"`
	EstimatedAvailabilities []struct {
		DeliveryOptions             []string `json:"deliveryOptions"`
		AvailabilityThresholdType   string   `json:"availabilityThresholdType"`
		AvailabilityThreshold       int      `json:"availabilityThreshold"`
		EstimatedAvailabilityStatus string   `json:"estimatedAvailabilityStatus"`
		EstimatedSoldQuantity       int      `json:"estimatedSoldQuantity"`
	} `json:"estimatedAvailabilities"`
	ShippingOptions []struct {
		ShippingServiceCode string `json:"shippingServiceCode"`
		Type                string `json:"type"`
		ShippingCost        struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"shippingCost"`
		QuantityUsedForEstimate       int       `json:"quantityUsedForEstimate"`
		MinEstimatedDeliveryDate      time.Time `json:"minEstimatedDeliveryDate"`
		MaxEstimatedDeliveryDate      time.Time `json:"maxEstimatedDeliveryDate"`
		ShipToLocationUsedForEstimate struct {
			Country string `json:"country"`
		} `json:"shipToLocationUsedForEstimate"`
		AdditionalShippingCostPerUnit struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"additionalShippingCostPerUnit"`
		ShippingCostType string `json:"shippingCostType"`
	} `json:"shippingOptions"`
	ShipToLocations struct {
		RegionIncluded []struct {
			RegionName string `json:"regionName"`
			RegionType string `json:"regionType"`
		} `json:"regionIncluded"`
		RegionExcluded []struct {
			RegionName string `json:"regionName"`
			RegionType string `json:"regionType"`
		} `json:"regionExcluded"`
	} `json:"shipToLocations"`
	ReturnTerms struct {
		ReturnsAccepted         bool   `json:"returnsAccepted"`
		RefundMethod            string `json:"refundMethod"`
		ReturnMethod            string `json:"returnMethod"`
		ReturnShippingCostPayer string `json:"returnShippingCostPayer"`
		ReturnPeriod            struct {
			Value int    `json:"value"`
			Unit  string `json:"unit"`
		} `json:"returnPeriod"`
		ReturnInstructions      string `json:"returnInstructions"`
		RestockingFeePercentage string `json:"restockingFeePercentage"`
	} `json:"returnTerms"`
	Taxes []struct {
		TaxJurisdiction struct {
			Region struct {
				RegionName string `json:"regionName"`
				RegionType string `json:"regionType"`
			} `json:"region"`
			TaxJurisdictionID string `json:"taxJurisdictionId"`
		} `json:"taxJurisdiction"`
		TaxType                  string `json:"taxType"`
		TaxPercentage            string `json:"taxPercentage"`
		ShippingAndHandlingTaxed bool   `json:"shippingAndHandlingTaxed"`
		IncludedInPrice          bool   `json:"includedInPrice"`
	} `json:"taxes"`
	LocalizedAspects []struct {
		Type  string `json:"type"`
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"localizedAspects"`
	QuantityLimitPerBuyer      int `json:"quantityLimitPerBuyer"`
	PrimaryProductReviewRating struct {
		ReviewCount      int    `json:"reviewCount"`
		AverageRating    string `json:"averageRating"`
		RatingHistograms []struct {
			Rating string `json:"rating"`
			Count  int    `json:"count"`
		} `json:"ratingHistograms"`
	} `json:"primaryProductReviewRating"`
	TopRatedBuyingExperience bool     `json:"topRatedBuyingExperience"`
	BuyingOptions            []string `json:"buyingOptions"`
	ItemAffiliateWebURL      string   `json:"itemAffiliateWebUrl"`
	ItemWebURL               string   `json:"itemWebUrl"`
	Description              string   `json:"description"`
	Product                  struct {
		AspectGroups []struct {
			LocalizedGroupName string `json:"localizedGroupName"`
			Aspects            []struct {
				LocalizedName   string   `json:"localizedName"`
				LocalizedValues []string `json:"localizedValues"`
			} `json:"aspects"`
		} `json:"aspectGroups"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Image       struct {
			ImageURL string `json:"imageUrl"`
		} `json:"image"`
		Gtins                       []string `json:"gtins"`
		Brand                       string   `json:"brand"`
		Mpns                        []string `json:"mpns"`
		AdditionalProductIdentities []struct {
			ProductIdentity []struct {
				IdentifierType  string `json:"identifierType"`
				IdentifierValue string `json:"identifierValue"`
			} `json:"productIdentity"`
		} `json:"additionalProductIdentities"`
	} `json:"product"`
	EnabledForGuestCheckout bool   `json:"enabledForGuestCheckout"`
	AdultOnly               bool   `json:"adultOnly"`
	CategoryID              string `json:"categoryId"`

	// Fields not present in the json sample provided by eBay:
	ItemEndDate       time.Time `json:"itemEndDate"`
	MinimumPriceToBid struct {
		Currency string `json:"currency"`
		Value    string `json:"value"`
	} `json:"minimumPriceToBid"`
	CurrentBidPrice struct {
		Currency string `json:"currency"`
		Value    string `json:"value"`
	} `json:"currentBidPrice"`
	UniqueBidderCount int `json:"uniqueBidderCount"`
}

// GetItem retrieves the details of a specific item.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/browse/resources/item/methods/getItem
func (s *BrowseService) GetItem(ctx context.Context, itemID string, opts ...Opt) (Item, error) {
	u := fmt.Sprintf("buy/browse/v1/item/%s?fieldgroups=PRODUCT", itemID)
	req, err := s.client.NewRequest(http.MethodGet, u, nil, opts...)
	if err != nil {
		return Item{}, err
	}
	var it Item
	return it, s.client.Do(ctx, req, &it)
}

// ItemsByGroup represents eBay items by group.
type ItemsByGroup struct {
	Items []struct {
		ItemID             string `json:"itemId"`
		SellerItemRevision string `json:"sellerItemRevision"`
		Title              string `json:"title"`
		ShortDescription   string `json:"shortDescription"`
		Price              struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"price"`
		CategoryPath string `json:"categoryPath"`
		Condition    string `json:"condition"`
		ConditionID  string `json:"conditionId"`
		ItemLocation struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"itemLocation"`
		Image struct {
			ImageURL string `json:"imageUrl"`
		} `json:"image"`
		Color       string    `json:"color"`
		Material    string    `json:"material"`
		Pattern     string    `json:"pattern"`
		SizeType    string    `json:"sizeType"`
		Brand       string    `json:"brand"`
		ItemEndDate time.Time `json:"itemEndDate"`
		Seller      struct {
			Username           string `json:"username"`
			FeedbackPercentage string `json:"feedbackPercentage"`
			FeedbackScore      int    `json:"feedbackScore"`
		} `json:"seller"`
		EstimatedAvailabilities []struct {
			DeliveryOptions             []string `json:"deliveryOptions"`
			AvailabilityThresholdType   string   `json:"availabilityThresholdType"`
			AvailabilityThreshold       int      `json:"availabilityThreshold"`
			EstimatedAvailabilityStatus string   `json:"estimatedAvailabilityStatus"`
			EstimatedSoldQuantity       int      `json:"estimatedSoldQuantity"`
		} `json:"estimatedAvailabilities"`
		ShippingOptions []struct {
			ShippingServiceCode string `json:"shippingServiceCode"`
			TrademarkSymbol     string `json:"trademarkSymbol,omitempty"`
			ShippingCarrierCode string `json:"shippingCarrierCode,omitempty"`
			Type                string `json:"type"`
			ShippingCost        struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"shippingCost"`
			QuantityUsedForEstimate       int       `json:"quantityUsedForEstimate"`
			MinEstimatedDeliveryDate      time.Time `json:"minEstimatedDeliveryDate"`
			MaxEstimatedDeliveryDate      time.Time `json:"maxEstimatedDeliveryDate"`
			ShipToLocationUsedForEstimate struct {
				Country string `json:"country"`
			} `json:"shipToLocationUsedForEstimate"`
			AdditionalShippingCostPerUnit struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"additionalShippingCostPerUnit"`
			ShippingCostType string `json:"shippingCostType"`
		} `json:"shippingOptions"`
		ShipToLocations struct {
			RegionIncluded []struct {
				RegionName string `json:"regionName"`
				RegionType string `json:"regionType"`
			} `json:"regionIncluded"`
			RegionExcluded []struct {
				RegionName string `json:"regionName"`
				RegionType string `json:"regionType"`
			} `json:"regionExcluded"`
		} `json:"shipToLocations"`
		ReturnTerms struct {
			ReturnsAccepted         bool   `json:"returnsAccepted"`
			RefundMethod            string `json:"refundMethod"`
			ReturnMethod            string `json:"returnMethod"`
			ReturnShippingCostPayer string `json:"returnShippingCostPayer"`
			ReturnPeriod            struct {
				Value int    `json:"value"`
				Unit  string `json:"unit"`
			} `json:"returnPeriod"`
		} `json:"returnTerms"`
		LocalizedAspects []struct {
			Type  string `json:"type"`
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"localizedAspects"`
		TopRatedBuyingExperience bool     `json:"topRatedBuyingExperience"`
		BuyingOptions            []string `json:"buyingOptions"`
		PrimaryItemGroup         struct {
			ItemGroupID    string `json:"itemGroupId"`
			ItemGroupType  string `json:"itemGroupType"`
			ItemGroupHref  string `json:"itemGroupHref"`
			ItemGroupTitle string `json:"itemGroupTitle"`
			ItemGroupImage struct {
				ImageURL string `json:"imageUrl"`
			} `json:"itemGroupImage"`
			ItemGroupAdditionalImages []struct {
				ImageURL string `json:"imageUrl"`
			} `json:"itemGroupAdditionalImages"`
		} `json:"primaryItemGroup"`
		EnabledForGuestCheckout bool   `json:"enabledForGuestCheckout"`
		AdultOnly               bool   `json:"adultOnly"`
		CategoryID              string `json:"categoryId"`
	} `json:"items"`
	CommonDescriptions []struct {
		Description string   `json:"description"`
		ItemIds     []string `json:"itemIds"`
	} `json:"commonDescriptions"`
}

// GetItemByGroupID retrieves the details of the individual items in an item group.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/browse/resources/item/methods/getItemsByItemGroup
func (s *BrowseService) GetItemByGroupID(ctx context.Context, groupID string, opts ...Opt) (ItemsByGroup, error) {
	u := fmt.Sprintf("buy/browse/v1/item/get_items_by_item_group?item_group_id=%s", groupID)
	req, err := s.client.NewRequest(http.MethodGet, u, nil, opts...)
	if err != nil {
		return ItemsByGroup{}, err
	}
	var it ItemsByGroup
	return it, s.client.Do(ctx, req, &it)
}

// CompatibilityProperty represents a product property.
type CompatibilityProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Compatibility represents an item compatibility.
type Compatibility struct {
	CompatibilityStatus string `json:"compatibilityStatus"`
	Warnings            []struct {
		Category     string   `json:"category"`
		Domain       string   `json:"domain"`
		ErrorID      int      `json:"errorId"`
		InputRefIds  []string `json:"inputRefIds"`
		LongMessage  string   `json:"longMessage"`
		Message      string   `json:"message"`
		OutputRefIds []string `json:"outputRefIds"`
		Parameters   []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"parameters"`
		Subdomain string `json:"subdomain"`
	} `json:"warnings"`
}

// Valid values for the "compatibilityStatus" compatibility field.
const (
	BrowseCheckComoatibilityCompatible    = "COMPATIBLE"
	BrowseCheckComoatibilityNotCompatible = "NOT_COMPATIBLE"
	BrowseCheckComoatibilityUndertermined = "UNDETERMINED"
)

// CheckCompatibility checks a product is compatible with the specified item.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/browse/resources/item/methods/checkCompatibility
func (s *BrowseService) CheckCompatibility(ctx context.Context, itemID, marketplaceID string, properties []CompatibilityProperty, opts ...Opt) (Compatibility, error) {
	type payload struct {
		CompatibilityProperties []CompatibilityProperty `json:"compatibilityProperties"`
	}
	pl := payload{properties}
	u := fmt.Sprintf("buy/browse/v1/item/%s/check_compatibility", itemID)
	opts = append(opts, OptBuyMarketplace(marketplaceID))
	req, err := s.client.NewRequest(http.MethodPost, u, &pl, opts...)
	if err != nil {
		return Compatibility{}, err
	}
	var c Compatibility
	return c, s.client.Do(ctx, req, &c)
}

// Search represents the result of an eBay search.
type Search struct {
	Href          string `json:"href"`
	Total         int    `json:"total"`
	Next          string `json:"next"`
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
	ItemSummaries []struct {
		ItemID string `json:"itemId"`
		Title  string `json:"title"`
		Image  struct {
			ImageURL string `json:"imageUrl"`
		} `json:"image"`
		Price struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"price"`
		ItemHref string `json:"itemHref"`
		Seller   struct {
			Username           string `json:"username"`
			FeedbackPercentage string `json:"feedbackPercentage"`
			FeedbackScore      int    `json:"feedbackScore"`
		} `json:"seller"`
		MarketingPrice struct {
			OriginalPrice struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"originalPrice"`
			DiscountPercentage string `json:"discountPercentage"`
			DiscountAmount     struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"discountAmount"`
		} `json:"marketingPrice"`
		Condition       string `json:"condition"`
		ConditionID     string `json:"conditionId"`
		ThumbnailImages []struct {
			ImageURL string `json:"imageUrl"`
		} `json:"thumbnailImages"`
		ShippingOptions []struct {
			ShippingCostType string `json:"shippingCostType"`
			ShippingCost     struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"shippingCost"`
		} `json:"shippingOptions"`
		BuyingOptions   []string `json:"buyingOptions"`
		CurrentBidPrice struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"currentBidPrice"`
		Epid         string `json:"epid"`
		ItemWebURL   string `json:"itemWebUrl"`
		ItemLocation struct {
			PostalCode string `json:"postalCode"`
			Country    string `json:"country"`
		} `json:"itemLocation"`
		Categories []struct {
			CategoryID string `json:"categoryId"`
		} `json:"categories"`
		AdditionalImages []struct {
			ImageURL string `json:"imageUrl"`
		} `json:"additionalImages"`
		AdultOnly bool `json:"adultOnly"`
	} `json:"itemSummaries"`
}

func optSearch(param string) func(v string) func(*http.Request) {
	return func(v string) func(*http.Request) {
		return func(req *http.Request) {
			query := req.URL.Query()
			query.Add(param, v)
			req.URL.RawQuery = query.Encode()
		}
	}
}

// Several query parameters to use with the Search method.

func OptBrowseSearch(v string) func(*http.Request) {
	return optSearch("q")(v)
}

func OptBrowseSearchGtin(v string) func(*http.Request) {
	return optSearch("gtin")(v)
}

func OptBrowseSearchCharityIDs(v string) func(*http.Request) {
	return optSearch("charity_ids")(v)
}

func OptBrowseSearchFieldgroups(v string) func(*http.Request) {
	return optSearch("fieldgroups")(v)
}

func OptBrowseSearchCompatibilityFilter(v string) func(*http.Request) {
	return optSearch("compatibility_filter")(v)
}

func OptBrowseSearchCategoryID(v string) func(*http.Request) {
	return optSearch("category_ids")(v)
}

func OptBrowseSearchFilter(v string) func(*http.Request) {
	return optSearch("filter")(v)
}

func OptBrowseSearchSort(v string) func(*http.Request) {
	return optSearch("sort")(v)
}

func OptBrowseSearchLimit(limit int) func(*http.Request) {
	return optSearch("limit")(strconv.Itoa(limit))
}

func OptBrowseSearchOffset(offset int) func(*http.Request) {
	return optSearch("offset")(strconv.Itoa(offset))
}

func OptBrowseSearchAspectFilter(v string) func(*http.Request) {
	return optSearch("aspect_filter")(v)
}

func OptBrowseSearchEPID(epid int) func(*http.Request) {
	return optSearch("epid")(strconv.Itoa(epid))
}

// Search searches for eBay items.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/browse/resources/item_summary/methods/search
func (s *BrowseService) Search(ctx context.Context, opts ...Opt) (Search, error) {
	u := "buy/browse/v1/item_summary/search"
	req, err := s.client.NewRequest(http.MethodGet, u, nil, opts...)
	if err != nil {
		return Search{}, err
	}
	var search Search
	return search, s.client.Do(ctx, req, &search)
}
