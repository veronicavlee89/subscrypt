package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Catzkorn/subscrypt/internal/subscription"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/shopspring/decimal"
)

func TestDatabaseConnection(t *testing.T) {

	t.Run("tests a successful database connection", func(t *testing.T) {
		_, err := NewDatabaseConnection(os.Getenv("DATABASE_CONN_STRING"))
		assertDatabaseError(t, err)
	})

	t.Run("tests a database connection failure", func(t *testing.T) {
		_, err := NewDatabaseConnection("gary the gopher")

		if err == nil {
			t.Errorf("connected to database that doesn't exist")
		}
	})
}

func TestRecordSubscriptionToDB(t *testing.T) {
	store, err := NewDatabaseConnection(os.Getenv("DATABASE_CONN_STRING"))
	assertDatabaseError(t, err)

	t.Run("adds a Netflix subscription", func(t *testing.T) {
		wantedSubscription := createTestSubscription("Netflix", "14.99", time.Date(2020, time.November, 29, 0, 0, 0, 0, time.UTC))
		subscription, err := store.RecordSubscription(wantedSubscription)
		assertDatabaseError(t, err)

		if subscription.ID == 0 {
			t.Errorf("Database did not return an ID, got %v want %v", 0, subscription.ID)
		}

		if subscription.Name != wantedSubscription.Name {
			t.Errorf("Database did not return correct subscription name, got %s want %s", subscription.Name, wantedSubscription.Name)
		}

		if !subscription.Amount.Equal(subscription.Amount) {
			t.Errorf("Database did not return correct amount, got %#v want %#v", subscription.Amount, wantedSubscription.Amount)
		}

		if !subscription.DateDue.Equal(wantedSubscription.DateDue) {
			t.Errorf("Database did not return correct subscription date, got %s want %s", subscription.DateDue, wantedSubscription.DateDue)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})

	t.Run("adds a climbing subscription", func(t *testing.T) {
		wantedSubscription := createTestSubscription("Reading Climbing Centre", "50.00", time.Date(2020, time.December, 30, 0, 0, 0, 0, time.UTC))

		subscription, err := store.RecordSubscription(wantedSubscription)
		assertDatabaseError(t, err)

		if subscription.ID == 0 {
			t.Errorf("Database did not return an ID, got %v want %v", 0, subscription.ID)
		}

		if subscription.Name != wantedSubscription.Name {
			t.Errorf("Database did not return correct subscription name, got %s want %s", subscription.Name, wantedSubscription.Name)
		}

		if !subscription.Amount.Equal(subscription.Amount) {
			t.Errorf("Database did not return correct amount, got %#v want %#v", subscription.Amount, wantedSubscription.Amount)
		}

		if !subscription.DateDue.Equal(wantedSubscription.DateDue) {
			t.Errorf("Database did not return correct subscription date, got %s want %s", subscription.DateDue, wantedSubscription.DateDue)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})

	t.Run("fails to add a subscription", func(t *testing.T) {
		emptySubscription := subscription.Subscription{}
		subscription, err := store.RecordSubscription(emptySubscription)
		assertDatabaseError(t, err)

		if subscription.Name != "" {
			t.Errorf("database retrieved a subscription when it was not meant to: %w", err)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})
}

func TestGetSubscriptionsFromDB(t *testing.T) {
	store, err := NewDatabaseConnection(os.Getenv("DATABASE_CONN_STRING"))
	assertDatabaseError(t, err)

	t.Run("gets all the subscriptions from the database", func(t *testing.T) {
		subscription := createTestSubscription("Amazon Prime", "7.99", time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC))

		wantedSubscription, err := store.RecordSubscription(subscription)
		assertDatabaseError(t, err)

		gotSubscriptions, err := store.GetSubscriptions()
		assertDatabaseError(t, err)

		if gotSubscriptions[0].ID != wantedSubscription.ID {
			t.Errorf("Database did not return an ID, got %v want %v", gotSubscriptions[0].ID, wantedSubscription.ID)
		}

		if gotSubscriptions[0].Name != wantedSubscription.Name {
			t.Errorf("Database did not return correct subscription name, got %s want %s", gotSubscriptions[0].Name, wantedSubscription.Name)
		}

		if !gotSubscriptions[0].Amount.Equal(wantedSubscription.Amount) {
			t.Errorf("Database did not return correct amount, got %#v want %#v", gotSubscriptions[0].Amount, wantedSubscription.Amount)
		}

		if !gotSubscriptions[0].DateDue.Equal(wantedSubscription.DateDue) {
			t.Errorf("Database did not return correct subscription date, got %s want %s", gotSubscriptions[0].DateDue, wantedSubscription.DateDue)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})

	t.Run("retrieves multiple subscriptions from the database", func(t *testing.T) {
		helloFreshSub := createTestSubscription("Hello Fresh", "80.00", time.Date(2020, time.November, 18, 0, 0, 0, 0, time.UTC))
		riverfordSub := createTestSubscription("Riverford", "180.00", time.Date(2020, time.December, 5, 0, 0, 0, 0, time.UTC))
		gymSub := createTestSubscription("PureGym", "34.99", time.Date(2020, time.December, 8, 0, 0, 0, 0, time.UTC))

		_, err := store.RecordSubscription(helloFreshSub)
		assertDatabaseError(t, err)
		_, err = store.RecordSubscription(riverfordSub)
		assertDatabaseError(t, err)
		_, err = store.RecordSubscription(gymSub)
		assertDatabaseError(t, err)

		gotSubscriptions, err := store.GetSubscriptions()
		assertDatabaseError(t, err)

		if len(gotSubscriptions) != 3 {
			t.Errorf("database did not return correct number of subscriptions, got %v want %v", len(gotSubscriptions), 3)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})

	t.Run("correctly handles retrieving from an empty database", func(t *testing.T) {
		gotSubscriptions, err := store.GetSubscriptions()
		assertDatabaseError(t, err)

		if len(gotSubscriptions) != 0 {
			t.Errorf("database retrieved subscriptions unexpectedly: %w", err)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})
}

func TestGetSubscriptionByIDFromDB(t *testing.T) {
	store, err := NewDatabaseConnection(os.Getenv("DATABASE_CONN_STRING"))
	assertDatabaseError(t, err)

	t.Run("returns subscription with given ID from DB", func(t *testing.T) {
		subscription := createTestSubscription("Graze Box", "20.00", time.Date(2021, time.February, 14, 0, 0, 0, 0, time.UTC))
		wantedSubscription, err := store.RecordSubscription(subscription)
		assertDatabaseError(t, err)

		gotSubscription, err := store.GetSubscription(wantedSubscription.ID)
		assertDatabaseError(t, err)

		if gotSubscription.ID != wantedSubscription.ID {
			t.Errorf("database did not return an ID, got %v want %v", 0, subscription.ID)
		}

		if gotSubscription.Name != wantedSubscription.Name {
			t.Errorf("database did not return correct subscription name, got %s want %s", subscription.Name, subscription.Name)
		}

		if !gotSubscription.Amount.Equal(wantedSubscription.Amount) {
			t.Errorf("database did not return correct amount, got %#v want %#v", subscription.Amount, subscription.Amount)
		}

		if !gotSubscription.DateDue.Equal(wantedSubscription.DateDue) {
			t.Errorf("database did not return correct subscription date, got %s want %s", subscription.DateDue, subscription.DateDue)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})

	t.Run("gets a specific subscription from the database", func(t *testing.T) {
		subscription := createTestSubscription("F1 TV", "8.95", time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC))

		wantedSubscription, err := store.RecordSubscription(subscription)
		assertDatabaseError(t, err)

		gotSubscription, err := store.GetSubscription(wantedSubscription.ID)
		assertDatabaseError(t, err)

		if gotSubscription.ID != wantedSubscription.ID {
			t.Errorf("Database did not return an ID, got %v want %v", gotSubscription.ID, wantedSubscription.ID)
		}

		if gotSubscription.Name != subscription.Name {
			t.Errorf("Database did not return correct subscription name, got %s want %s", gotSubscription.Name, wantedSubscription.Name)
		}

		if !gotSubscription.Amount.Equal(subscription.Amount) {
			t.Errorf("Database did not return correct amount, got %#v want %#v", gotSubscription.Amount, wantedSubscription.Amount)
		}

		if !gotSubscription.DateDue.Equal(subscription.DateDue) {
			t.Errorf("Database did not return correct subscription date, got %s want %s", gotSubscription.DateDue, wantedSubscription.DateDue)
		}
		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})

	t.Run("returns subscription with given ID from DB with multiple records", func(t *testing.T) {
		helloFreshSub := createTestSubscription("Hello Fresh", "80.00", time.Date(2020, time.November, 18, 0, 0, 0, 0, time.UTC))
		riverfordSub := createTestSubscription("Riverford", "180.00", time.Date(2020, time.December, 5, 0, 0, 0, 0, time.UTC))
		gymSub := createTestSubscription("PureGym", "34.99", time.Date(2020, time.December, 8, 0, 0, 0, 0, time.UTC))

		_, err := store.RecordSubscription(helloFreshSub)
		assertDatabaseError(t, err)
		wantedSubscription, err := store.RecordSubscription(riverfordSub)
		assertDatabaseError(t, err)
		_, err = store.RecordSubscription(gymSub)
		assertDatabaseError(t, err)

		gotSubscription, err := store.GetSubscription(wantedSubscription.ID)
		assertDatabaseError(t, err)

		if gotSubscription.ID != wantedSubscription.ID {
			t.Errorf("Database did not return an ID, got %v want %v", 0, riverfordSub.ID)
		}

		if gotSubscription.Name != wantedSubscription.Name {
			t.Errorf("Database did not return correct subscription name, got %s want %s", gotSubscription.Name, riverfordSub.Name)
		}

		if !gotSubscription.Amount.Equal(wantedSubscription.Amount) {
			t.Errorf("Database did not return correct amount, got %#v want %#v", gotSubscription.Amount, riverfordSub.Amount)
		}

		if !gotSubscription.DateDue.Equal(wantedSubscription.DateDue) {
			t.Errorf("Database did not return correct subscription date, got %s want %s", gotSubscription.DateDue, riverfordSub.DateDue)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})

	t.Run("returns nil if a subscription with the given ID does not exist in the DB", func(t *testing.T) {
		gotSubscription, err := store.GetSubscription(2)
		assertDatabaseError(t, err)

		if gotSubscription != nil {
			t.Errorf("database returned value for non-existent subscription, got %v want %v", gotSubscription, nil)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})
}

func TestDeletingSubscriptionFromDB(t *testing.T) {
	store, err := NewDatabaseConnection(os.Getenv("DATABASE_CONN_STRING"))
	assertDatabaseError(t, err)

	t.Run("deletes a subscription from the database", func(t *testing.T) {
		subscription := createTestSubscription("BMC", "11.99", time.Date(2020, time.December, 1, 0, 0, 0, 0, time.UTC))

		gotSubscription, err := store.RecordSubscription(subscription)
		assertDatabaseError(t, err)

		subscriptionID := gotSubscription.ID

		err = store.DeleteSubscription(subscriptionID)
		assertDatabaseError(t, err)

		gotSubscription, err = store.GetSubscription(subscriptionID)
		assertDatabaseError(t, err)

		if gotSubscription != nil {
			t.Errorf("database did not delete subscription, got subscription name %v, wanted no subscriptions", gotSubscription.Name)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})

	t.Run("attempts to delete a subscription by an invalid ID", func(t *testing.T) {
		err := store.DeleteSubscription(0)
		if err == nil {
			t.Errorf("deleting invalid subscription did not error")
		}
	})

	t.Run("deletes both instances of a subscription", func(t *testing.T) {
		subscription := createTestSubscription("Apple TV", "7.99", time.Date(2020, time.December, 15, 0, 0, 0, 0, time.UTC))

		_, err := store.RecordSubscription(subscription)
		assertDatabaseError(t, err)
		gotSubscription, err := store.RecordSubscription(subscription)
		assertDatabaseError(t, err)

		err = store.DeleteSubscription(gotSubscription.ID)
		assertDatabaseError(t, err)

		subscriptions, err := store.GetSubscriptions()
		assertDatabaseError(t, err)

		if len(subscriptions) != 0 {
			t.Errorf("database did not delete all instances of the subscription, got %v, wanted no subscriptions", subscriptions)
		}

		err = clearSubscriptionsTable()
		assertDatabaseError(t, err)
	})
}

func TestUserprofilesDatabase(t *testing.T) {
	usersName := "Gary Gopher"
	usersEmail := "gary@gopher.com"

	store, err := NewDatabaseConnection(os.Getenv("DATABASE_CONN_STRING"))
	assertDatabaseError(t, err)

	t.Run("add name and email", func(t *testing.T) {
		returnedDetails, err := store.RecordUserDetails(usersName, usersEmail)
		assertDatabaseError(t, err)

		if returnedDetails == nil {
			t.Fatalf("no user details recorded: %v", err)
		}

		if returnedDetails.Name != usersName {
			t.Errorf("incorrect user name returned got %v want %v", returnedDetails.Name, usersName)
		}

		if returnedDetails.Email != usersEmail {
			t.Errorf("incorrect email returned got %v want %v", returnedDetails.Name, usersEmail)
		}
		err = clearUsersTable()
		assertDatabaseError(t, err)
	})

	t.Run("get name and email from database", func(t *testing.T) {
		_, err := store.RecordUserDetails(usersName, usersEmail)
		assertDatabaseError(t, err)

		gotDetails, err := store.GetUserDetails()
		assertDatabaseError(t, err)

		if gotDetails.Name != usersName {
			t.Errorf("incorrect name retrieved got %v want %v", gotDetails.Name, usersName)
		}

		if gotDetails.Email != usersEmail {
			t.Errorf("incorrect email retrieved got %v want %v", gotDetails.Email, usersEmail)
		}

		err = clearUsersTable()
		assertDatabaseError(t, err)
	})

	t.Run("update name and email", func(t *testing.T) {
		_, err := store.RecordUserDetails(usersName, usersEmail)
		assertDatabaseError(t, err)

		updatedName := "Gwen Gopher"
		updatedEmail := "gwen@gopher.com"

		updatedUser, err := store.RecordUserDetails(updatedName, updatedEmail)
		assertDatabaseError(t, err)

		if updatedUser.Name != updatedName {
			t.Errorf("incorrect name retrieved got %v want %v", updatedUser.Name, updatedName)
		}

		if updatedUser.Email != updatedEmail {
			t.Errorf("incorrect name retrieved got %v want %v", updatedUser.Email, updatedEmail)
		}
		err = clearUsersTable()
		assertDatabaseError(t, err)
	})
}

func createTestSubscription(name string, price string, date time.Time) subscription.Subscription {
	amount, _ := decimal.NewFromString(price)
	subscription := subscription.Subscription{
		Name:    name,
		Amount:  amount,
		DateDue: date,
	}
	return subscription
}

func clearSubscriptionsTable() error {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_CONN_STRING"))
	if err != nil {
		return fmt.Errorf("unexpected connection error: %w", err)
	}
	_, err = db.ExecContext(context.Background(), "TRUNCATE TABLE subscriptions;")

	return err
}

func clearUsersTable() error {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_CONN_STRING"))
	if err != nil {
		return fmt.Errorf("unexpected connection error: %w", err)
	}
	_, err = db.ExecContext(context.Background(), "TRUNCATE TABLE users;")

	return err
}

func assertDatabaseError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected database error: %v", err)
	}
}
