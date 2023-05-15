// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package graphql_generated

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/99designs/gqlgen/graphql"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/vektah/gqlparser/v2/ast"
)

// region    ************************** generated!.gotpl **************************

type FiatDepositResponseResolver interface {
	TxID(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error)
	ClientID(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error)
	TxTimestamp(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error)
	Balance(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error)
	LastTx(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error)
	Currency(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error)
}
type FiatExchangeOfferResponseResolver interface {
	DebitAmount(ctx context.Context, obj *models.HTTPFiatExchangeOfferResponse) (float64, error)
}

type FiatDepositRequestResolver interface {
	Amount(ctx context.Context, obj *models.HTTPDepositCurrencyRequest, data float64) error
}
type FiatExchangeOfferRequestResolver interface {
	SourceAmount(ctx context.Context, obj *models.HTTPFiatExchangeOfferRequest, data float64) error
}

// endregion ************************** generated!.gotpl **************************

// region    ***************************** args.gotpl *****************************

// endregion ***************************** args.gotpl *****************************

// region    ************************** directives.gotpl **************************

// endregion ************************** directives.gotpl **************************

// region    **************************** field.gotpl *****************************

func (ec *executionContext) _FiatDepositResponse_txId(ctx context.Context, field graphql.CollectedField, obj *postgres.FiatAccountTransferResult) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatDepositResponse_txId(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return ec.resolvers.FiatDepositResponse().TxID(rctx, obj)
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNString2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatDepositResponse_txId(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatDepositResponse",
		Field:      field,
		IsMethod:   true,
		IsResolver: true,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type String does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatDepositResponse_clientId(ctx context.Context, field graphql.CollectedField, obj *postgres.FiatAccountTransferResult) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatDepositResponse_clientId(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return ec.resolvers.FiatDepositResponse().ClientID(rctx, obj)
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNString2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatDepositResponse_clientId(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatDepositResponse",
		Field:      field,
		IsMethod:   true,
		IsResolver: true,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type String does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatDepositResponse_txTimestamp(ctx context.Context, field graphql.CollectedField, obj *postgres.FiatAccountTransferResult) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatDepositResponse_txTimestamp(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return ec.resolvers.FiatDepositResponse().TxTimestamp(rctx, obj)
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNString2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatDepositResponse_txTimestamp(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatDepositResponse",
		Field:      field,
		IsMethod:   true,
		IsResolver: true,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type String does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatDepositResponse_balance(ctx context.Context, field graphql.CollectedField, obj *postgres.FiatAccountTransferResult) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatDepositResponse_balance(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return ec.resolvers.FiatDepositResponse().Balance(rctx, obj)
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNString2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatDepositResponse_balance(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatDepositResponse",
		Field:      field,
		IsMethod:   true,
		IsResolver: true,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type String does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatDepositResponse_lastTx(ctx context.Context, field graphql.CollectedField, obj *postgres.FiatAccountTransferResult) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatDepositResponse_lastTx(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return ec.resolvers.FiatDepositResponse().LastTx(rctx, obj)
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNString2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatDepositResponse_lastTx(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatDepositResponse",
		Field:      field,
		IsMethod:   true,
		IsResolver: true,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type String does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatDepositResponse_currency(ctx context.Context, field graphql.CollectedField, obj *postgres.FiatAccountTransferResult) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatDepositResponse_currency(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return ec.resolvers.FiatDepositResponse().Currency(rctx, obj)
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNString2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatDepositResponse_currency(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatDepositResponse",
		Field:      field,
		IsMethod:   true,
		IsResolver: true,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type String does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatExchangeOfferResponse_priceQuote(ctx context.Context, field graphql.CollectedField, obj *models.HTTPFiatExchangeOfferResponse) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatExchangeOfferResponse_priceQuote(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return obj.PriceQuote, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(models.PriceQuote)
	fc.Result = res
	return ec.marshalNPriceQuote2githubᚗcomᚋsurahmanᚋFTeXᚋpkgᚋmodelsᚐPriceQuote(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatExchangeOfferResponse_priceQuote(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatExchangeOfferResponse",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "ClientID":
				return ec.fieldContext_PriceQuote_ClientID(ctx, field)
			case "SourceAcc":
				return ec.fieldContext_PriceQuote_SourceAcc(ctx, field)
			case "DestinationAcc":
				return ec.fieldContext_PriceQuote_DestinationAcc(ctx, field)
			case "Rate":
				return ec.fieldContext_PriceQuote_Rate(ctx, field)
			case "Amount":
				return ec.fieldContext_PriceQuote_Amount(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type PriceQuote", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatExchangeOfferResponse_debitAmount(ctx context.Context, field graphql.CollectedField, obj *models.HTTPFiatExchangeOfferResponse) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatExchangeOfferResponse_debitAmount(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return ec.resolvers.FiatExchangeOfferResponse().DebitAmount(rctx, obj)
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(float64)
	fc.Result = res
	return ec.marshalNFloat2float64(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatExchangeOfferResponse_debitAmount(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatExchangeOfferResponse",
		Field:      field,
		IsMethod:   true,
		IsResolver: true,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type Float does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatExchangeOfferResponse_offerID(ctx context.Context, field graphql.CollectedField, obj *models.HTTPFiatExchangeOfferResponse) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatExchangeOfferResponse_offerID(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return obj.OfferID, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNString2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatExchangeOfferResponse_offerID(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatExchangeOfferResponse",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type String does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatExchangeOfferResponse_expires(ctx context.Context, field graphql.CollectedField, obj *models.HTTPFiatExchangeOfferResponse) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatExchangeOfferResponse_expires(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return obj.Expires, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(int64)
	fc.Result = res
	return ec.marshalNInt642int64(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatExchangeOfferResponse_expires(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatExchangeOfferResponse",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type Int64 does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatOpenAccountResponse_clientID(ctx context.Context, field graphql.CollectedField, obj *models.FiatOpenAccountResponse) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatOpenAccountResponse_clientID(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return obj.ClientID, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNString2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatOpenAccountResponse_clientID(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatOpenAccountResponse",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type String does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _FiatOpenAccountResponse_currency(ctx context.Context, field graphql.CollectedField, obj *models.FiatOpenAccountResponse) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_FiatOpenAccountResponse_currency(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (interface{}, error) {
		ctx = rctx // use context from middleware stack in children
		return obj.Currency, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNString2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_FiatOpenAccountResponse_currency(ctx context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "FiatOpenAccountResponse",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type String does not have child fields")
		},
	}
	return fc, nil
}

// endregion **************************** field.gotpl *****************************

// region    **************************** input.gotpl *****************************

func (ec *executionContext) unmarshalInputFiatDepositRequest(ctx context.Context, obj interface{}) (models.HTTPDepositCurrencyRequest, error) {
	var it models.HTTPDepositCurrencyRequest
	asMap := map[string]interface{}{}
	for k, v := range obj.(map[string]interface{}) {
		asMap[k] = v
	}

	fieldsInOrder := [...]string{"amount", "currency"}
	for _, k := range fieldsInOrder {
		v, ok := asMap[k]
		if !ok {
			continue
		}
		switch k {
		case "amount":
			var err error

			ctx := graphql.WithPathContext(ctx, graphql.NewPathWithField("amount"))
			data, err := ec.unmarshalNFloat2float64(ctx, v)
			if err != nil {
				return it, err
			}
			if err = ec.resolvers.FiatDepositRequest().Amount(ctx, &it, data); err != nil {
				return it, err
			}
		case "currency":
			var err error

			ctx := graphql.WithPathContext(ctx, graphql.NewPathWithField("currency"))
			data, err := ec.unmarshalNString2string(ctx, v)
			if err != nil {
				return it, err
			}
			it.Currency = data
		}
	}

	return it, nil
}

func (ec *executionContext) unmarshalInputFiatExchangeOfferRequest(ctx context.Context, obj interface{}) (models.HTTPFiatExchangeOfferRequest, error) {
	var it models.HTTPFiatExchangeOfferRequest
	asMap := map[string]interface{}{}
	for k, v := range obj.(map[string]interface{}) {
		asMap[k] = v
	}

	fieldsInOrder := [...]string{"sourceCurrency", "destinationCurrency", "sourceAmount"}
	for _, k := range fieldsInOrder {
		v, ok := asMap[k]
		if !ok {
			continue
		}
		switch k {
		case "sourceCurrency":
			var err error

			ctx := graphql.WithPathContext(ctx, graphql.NewPathWithField("sourceCurrency"))
			data, err := ec.unmarshalNString2string(ctx, v)
			if err != nil {
				return it, err
			}
			it.SourceCurrency = data
		case "destinationCurrency":
			var err error

			ctx := graphql.WithPathContext(ctx, graphql.NewPathWithField("destinationCurrency"))
			data, err := ec.unmarshalNString2string(ctx, v)
			if err != nil {
				return it, err
			}
			it.DestinationCurrency = data
		case "sourceAmount":
			var err error

			ctx := graphql.WithPathContext(ctx, graphql.NewPathWithField("sourceAmount"))
			data, err := ec.unmarshalNFloat2float64(ctx, v)
			if err != nil {
				return it, err
			}
			if err = ec.resolvers.FiatExchangeOfferRequest().SourceAmount(ctx, &it, data); err != nil {
				return it, err
			}
		}
	}

	return it, nil
}

// endregion **************************** input.gotpl *****************************

// region    ************************** interface.gotpl ***************************

// endregion ************************** interface.gotpl ***************************

// region    **************************** object.gotpl ****************************

var fiatDepositResponseImplementors = []string{"FiatDepositResponse"}

func (ec *executionContext) _FiatDepositResponse(ctx context.Context, sel ast.SelectionSet, obj *postgres.FiatAccountTransferResult) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, fiatDepositResponseImplementors)
	out := graphql.NewFieldSet(fields)
	var invalids uint32
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("FiatDepositResponse")
		case "txId":
			field := field

			innerFunc := func(ctx context.Context) (res graphql.Marshaler) {
				defer func() {
					if r := recover(); r != nil {
						ec.Error(ctx, ec.Recover(ctx, r))
					}
				}()
				res = ec._FiatDepositResponse_txId(ctx, field, obj)
				if res == graphql.Null {
					atomic.AddUint32(&invalids, 1)
				}
				return res
			}

			out.Concurrently(i, func() graphql.Marshaler {
				return innerFunc(ctx)

			})
		case "clientId":
			field := field

			innerFunc := func(ctx context.Context) (res graphql.Marshaler) {
				defer func() {
					if r := recover(); r != nil {
						ec.Error(ctx, ec.Recover(ctx, r))
					}
				}()
				res = ec._FiatDepositResponse_clientId(ctx, field, obj)
				if res == graphql.Null {
					atomic.AddUint32(&invalids, 1)
				}
				return res
			}

			out.Concurrently(i, func() graphql.Marshaler {
				return innerFunc(ctx)

			})
		case "txTimestamp":
			field := field

			innerFunc := func(ctx context.Context) (res graphql.Marshaler) {
				defer func() {
					if r := recover(); r != nil {
						ec.Error(ctx, ec.Recover(ctx, r))
					}
				}()
				res = ec._FiatDepositResponse_txTimestamp(ctx, field, obj)
				if res == graphql.Null {
					atomic.AddUint32(&invalids, 1)
				}
				return res
			}

			out.Concurrently(i, func() graphql.Marshaler {
				return innerFunc(ctx)

			})
		case "balance":
			field := field

			innerFunc := func(ctx context.Context) (res graphql.Marshaler) {
				defer func() {
					if r := recover(); r != nil {
						ec.Error(ctx, ec.Recover(ctx, r))
					}
				}()
				res = ec._FiatDepositResponse_balance(ctx, field, obj)
				if res == graphql.Null {
					atomic.AddUint32(&invalids, 1)
				}
				return res
			}

			out.Concurrently(i, func() graphql.Marshaler {
				return innerFunc(ctx)

			})
		case "lastTx":
			field := field

			innerFunc := func(ctx context.Context) (res graphql.Marshaler) {
				defer func() {
					if r := recover(); r != nil {
						ec.Error(ctx, ec.Recover(ctx, r))
					}
				}()
				res = ec._FiatDepositResponse_lastTx(ctx, field, obj)
				if res == graphql.Null {
					atomic.AddUint32(&invalids, 1)
				}
				return res
			}

			out.Concurrently(i, func() graphql.Marshaler {
				return innerFunc(ctx)

			})
		case "currency":
			field := field

			innerFunc := func(ctx context.Context) (res graphql.Marshaler) {
				defer func() {
					if r := recover(); r != nil {
						ec.Error(ctx, ec.Recover(ctx, r))
					}
				}()
				res = ec._FiatDepositResponse_currency(ctx, field, obj)
				if res == graphql.Null {
					atomic.AddUint32(&invalids, 1)
				}
				return res
			}

			out.Concurrently(i, func() graphql.Marshaler {
				return innerFunc(ctx)

			})
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}
	out.Dispatch()
	if invalids > 0 {
		return graphql.Null
	}
	return out
}

var fiatExchangeOfferResponseImplementors = []string{"FiatExchangeOfferResponse"}

func (ec *executionContext) _FiatExchangeOfferResponse(ctx context.Context, sel ast.SelectionSet, obj *models.HTTPFiatExchangeOfferResponse) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, fiatExchangeOfferResponseImplementors)
	out := graphql.NewFieldSet(fields)
	var invalids uint32
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("FiatExchangeOfferResponse")
		case "priceQuote":

			out.Values[i] = ec._FiatExchangeOfferResponse_priceQuote(ctx, field, obj)

			if out.Values[i] == graphql.Null {
				atomic.AddUint32(&invalids, 1)
			}
		case "debitAmount":
			field := field

			innerFunc := func(ctx context.Context) (res graphql.Marshaler) {
				defer func() {
					if r := recover(); r != nil {
						ec.Error(ctx, ec.Recover(ctx, r))
					}
				}()
				res = ec._FiatExchangeOfferResponse_debitAmount(ctx, field, obj)
				if res == graphql.Null {
					atomic.AddUint32(&invalids, 1)
				}
				return res
			}

			out.Concurrently(i, func() graphql.Marshaler {
				return innerFunc(ctx)

			})
		case "offerID":

			out.Values[i] = ec._FiatExchangeOfferResponse_offerID(ctx, field, obj)

			if out.Values[i] == graphql.Null {
				atomic.AddUint32(&invalids, 1)
			}
		case "expires":

			out.Values[i] = ec._FiatExchangeOfferResponse_expires(ctx, field, obj)

			if out.Values[i] == graphql.Null {
				atomic.AddUint32(&invalids, 1)
			}
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}
	out.Dispatch()
	if invalids > 0 {
		return graphql.Null
	}
	return out
}

var fiatOpenAccountResponseImplementors = []string{"FiatOpenAccountResponse"}

func (ec *executionContext) _FiatOpenAccountResponse(ctx context.Context, sel ast.SelectionSet, obj *models.FiatOpenAccountResponse) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, fiatOpenAccountResponseImplementors)
	out := graphql.NewFieldSet(fields)
	var invalids uint32
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("FiatOpenAccountResponse")
		case "clientID":

			out.Values[i] = ec._FiatOpenAccountResponse_clientID(ctx, field, obj)

			if out.Values[i] == graphql.Null {
				invalids++
			}
		case "currency":

			out.Values[i] = ec._FiatOpenAccountResponse_currency(ctx, field, obj)

			if out.Values[i] == graphql.Null {
				invalids++
			}
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}
	out.Dispatch()
	if invalids > 0 {
		return graphql.Null
	}
	return out
}

// endregion **************************** object.gotpl ****************************

// region    ***************************** type.gotpl *****************************

func (ec *executionContext) unmarshalNFiatDepositRequest2githubᚗcomᚋsurahmanᚋFTeXᚋpkgᚋmodelsᚐHTTPDepositCurrencyRequest(ctx context.Context, v interface{}) (models.HTTPDepositCurrencyRequest, error) {
	res, err := ec.unmarshalInputFiatDepositRequest(ctx, v)
	return res, graphql.ErrorOnPath(ctx, err)
}

func (ec *executionContext) marshalNFiatDepositResponse2githubᚗcomᚋsurahmanᚋFTeXᚋpkgᚋpostgresᚐFiatAccountTransferResult(ctx context.Context, sel ast.SelectionSet, v postgres.FiatAccountTransferResult) graphql.Marshaler {
	return ec._FiatDepositResponse(ctx, sel, &v)
}

func (ec *executionContext) marshalNFiatDepositResponse2ᚖgithubᚗcomᚋsurahmanᚋFTeXᚋpkgᚋpostgresᚐFiatAccountTransferResult(ctx context.Context, sel ast.SelectionSet, v *postgres.FiatAccountTransferResult) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._FiatDepositResponse(ctx, sel, v)
}

func (ec *executionContext) marshalNFiatOpenAccountResponse2githubᚗcomᚋsurahmanᚋFTeXᚋpkgᚋmodelsᚐFiatOpenAccountResponse(ctx context.Context, sel ast.SelectionSet, v models.FiatOpenAccountResponse) graphql.Marshaler {
	return ec._FiatOpenAccountResponse(ctx, sel, &v)
}

func (ec *executionContext) marshalNFiatOpenAccountResponse2ᚖgithubᚗcomᚋsurahmanᚋFTeXᚋpkgᚋmodelsᚐFiatOpenAccountResponse(ctx context.Context, sel ast.SelectionSet, v *models.FiatOpenAccountResponse) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._FiatOpenAccountResponse(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
