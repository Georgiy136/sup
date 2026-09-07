package redis

import (
	"context"
	"fmt"
	"strconv"

	accesspolicy "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/models/access_policy"

	jsoniter "github.com/json-iterator/go"
)

type AccessPolicyRepository struct {
	*RedisClient
}

func NewAccessPolicyRepository(client *RedisClient) *AccessPolicyRepository {
	return &AccessPolicyRepository{client}
}

func (r *AccessPolicyRepository) AddAccessPolicyForActionGroup(ctx context.Context, resource accesspolicy.ResourceCompositeKey, groupID int64) error {
	key := AccessPolicyForActionGroupKey(resource)

	err := r.rCli.SAdd(ctx, key, strconv.FormatInt(groupID, 10)).Err()
	if err != nil {
		return fmt.Errorf("can't add access policy action group in redis by key: %s, err: %w", key, err)
	}

	return nil
}

func (r *AccessPolicyRepository) DelAccessPolicyForActionGroup(ctx context.Context, resource accesspolicy.ResourceCompositeKey, groupID int64) error {
	key := AccessPolicyForActionGroupKey(resource)

	err := r.rCli.SRem(ctx, key, strconv.FormatInt(groupID, 10)).Err()
	if err != nil {
		return fmt.Errorf("can't del access policy action group in redis by key: %s, err: %w", key, err)
	}

	return nil
}

func (r *AccessPolicyRepository) AddAccessPolicyForExternalAction(ctx context.Context, resource accesspolicy.ResourceCompositeKey, externalAction string) error {
	key := AccessPolicyForExternalActionsKey(resource)

	err := r.rCli.SAdd(ctx, key, externalAction).Err()
	if err != nil {
		return fmt.Errorf("can't add access policy external actions in redis by key: %s, err: %w", key, err)
	}

	return nil
}

func (r *AccessPolicyRepository) DelAccessPolicyForExternalAction(ctx context.Context, resource accesspolicy.ResourceCompositeKey, externalAction string) error {
	key := AccessPolicyForExternalActionsKey(resource)

	err := r.rCli.SRem(ctx, key, externalAction).Err()
	if err != nil {
		return fmt.Errorf("can't del access policy external actions in redis by key: %s, err: %w", key, err)
	}

	return nil
}

func (r *AccessPolicyRepository) AddExternalActionToTypeAction(ctx context.Context, typeAction string, externalActionWithCategory accesspolicy.ExternalActionWithCategory) error {
	key := ExternalActionsToTypeActionKey(typeAction)

	value, err := jsoniter.Marshal(externalActionWithCategory)
	if err != nil {
		return fmt.Errorf("can't marshal external action with category: %w", err)
	}

	err = r.rCli.SAdd(ctx, key, value).Err()
	if err != nil {
		return fmt.Errorf("can't add external action to type action in redis by key: %s, err: %w", key, err)
	}

	return nil
}

func (r *AccessPolicyRepository) DelExternalActionToTypeAction(ctx context.Context, typeAction string, externalActionWithCategory accesspolicy.ExternalActionWithCategory) error {
	key := ExternalActionsToTypeActionKey(typeAction)

	value, err := jsoniter.Marshal(externalActionWithCategory)
	if err != nil {
		return fmt.Errorf("can't marshal external action with category: %w", err)
	}

	err = r.rCli.SRem(ctx, key, value).Err()
	if err != nil {
		return fmt.Errorf("can't del external action to type action in redis by key: %s, err: %w", key, err)
	}

	return nil
}

func (r *AccessPolicyRepository) AddGroupIdToTypeAction(ctx context.Context, typeAction string, groupId accesspolicy.GroupIdWithCategory) error {
	key := GroupIdsToTypeActionKey(typeAction)

	value, err := jsoniter.Marshal(groupId)
	if err != nil {
		return fmt.Errorf("can't marshal group id with category: %w", err)
	}

	err = r.rCli.SAdd(ctx, key, value).Err()
	if err != nil {
		return fmt.Errorf("can't add group id to type action in redis by key: %s, err: %w", key, err)
	}

	return nil
}

func (r *AccessPolicyRepository) DelGroupIdToTypeAction(ctx context.Context, typeAction string, groupId accesspolicy.GroupIdWithCategory) error {
	key := GroupIdsToTypeActionKey(typeAction)

	value, err := jsoniter.Marshal(groupId)
	if err != nil {
		return fmt.Errorf("can't marshal group id with category: %w", err)
	}

	err = r.rCli.SRem(ctx, key, value).Err()
	if err != nil {
		return fmt.Errorf("can't del group id to type action in redis by key: %s, err: %w", key, err)
	}

	return nil
}
