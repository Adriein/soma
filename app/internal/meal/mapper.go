package meal

import "github.com/adriein/soma/app/pkg/vendor"

func ToDomain(fsMeal *vendor.DiaryMeal) *Meal {
	return &Meal{
		ID:                 fsMeal.ID,
		Calcium:            fsMeal.Calcium,
		Calories:           fsMeal.Calories,
		Carbohydrate:       fsMeal.Carbohydrate,
		Cholesterol:        fsMeal.Cholesterol,
		DateInt:            fsMeal.DateInt,
		Fat:                fsMeal.Fat,
		Fiber:              fsMeal.Fiber,
		Description:        fsMeal.Description,
		Name:               fsMeal.Name,
		FoodID:             fsMeal.FoodID,
		Iron:               fsMeal.Iron,
		Meal:               fsMeal.Meal,
		MonounsaturatedFat: fsMeal.MonounsaturatedFat,
		NumberOfUnits:      fsMeal.NumberOfUnits,
		PolyunsaturatedFat: fsMeal.PolyunsaturatedFat,
		Potassium:          fsMeal.Potassium,
		Protein:            fsMeal.Protein,
		SaturatedFat:       fsMeal.SaturatedFat,
		ServingID:          fsMeal.ServingID,
		Sodium:             fsMeal.Sodium,
		Sugar:              fsMeal.Sugar,
		VitaminA:           fsMeal.VitaminA,
		VitaminC:           fsMeal.VitaminC,
	}
}
