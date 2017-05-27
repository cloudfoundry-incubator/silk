package database_test

import (
	"database/sql"
	"errors"
	"time"

	database "code.cloudfoundry.org/silk/controller/database"
	"code.cloudfoundry.org/silk/controller/database/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deleter", func() {
	var (
		group          *database.Deleter
		expectedResult sql.Result
		tx             *fakes.Transaction
	)

	BeforeEach(func() {
		group = &database.Deleter{}
		tx = &fakes.Transaction{}

	})
	Describe("Delete", func() {
		BeforeEach(func() {
			expectedResult = &fakes.SqlResult{}
			tx.ExecContextReturns(expectedResult, nil)
			tx.RebindReturns("rebinded query")
		})
		It("deletes from the subnets table", func() {
			result, err := group.Delete(tx, "underlay", 4*time.Second)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expectedResult))

			Expect(tx.RebindCallCount()).To(Equal(1))
			query := tx.RebindArgsForCall(0)
			Expect(query).To(Equal("DELETE FROM subnets WHERE underlay_ip = ?"))

			Expect(tx.ExecContextCallCount()).To(Equal(1))
			_, query, underlayIP := tx.ExecContextArgsForCall(0)
			Expect(query).To(Equal("rebinded query"))
			Expect(underlayIP).To(ConsistOf("underlay"))
		})

		Context("when ExecContext returns an error", func() {
			BeforeEach(func() {
				tx.ExecContextReturns(nil, errors.New("banana"))
			})
			It("returns the error", func() {
				_, err := group.Delete(tx, "underlay", 4*time.Second)
				Expect(err).To(MatchError("banana"))

			})
		})
	})
})
