module bluelink::bluelink {
    use sui::object::{Self, ID, UID};
    use sui::tx_context::{Self, TxContext};
    use sui::transfer;
    use sui::coin::{Self, Coin};
    use sui::sui::SUI;
    use sui::balance::{Self, Balance};
    use sui::event;
    use std::string::{Self, String};
    use std::vector;

    // Error codes
    const EInsufficientFunds: u64 = 1;
    const ENotProjectCreator: u64 = 2;
    const EInsufficientBalance: u64 = 3;

    // Struct definitions
    struct Project has key, store {
        id: UID,
        creator: address,
        name: String,
        description: String,
        funding_goal: u64,
        total_raised: Balance<SUI>,
        donor_count: u64,
    }

    struct DonationReceipt has key, store {
        id: UID,
        project_id: ID,
        donor: address,
        amount: u64,
    }

    // Events
    struct ProjectCreated has copy, drop {
        project_id: ID,
        creator: address,
        name: String,
        funding_goal: u64,
    }

    struct DonationMade has copy, drop {
        project_id: ID,
        donor: address,
        amount: u64,
        receipt_id: ID,
    }

    struct FundsWithdrawn has copy, drop {
        project_id: ID,
        creator: address,
        amount: u64,
    }

    // Entry functions
    public entry fun create_project(
        name: vector<u8>,
        description: vector<u8>,
        funding_goal: u64,
        ctx: &mut TxContext
    ) {
        let creator = tx_context::sender(ctx);
        let project_id = object::new(ctx);
        let project_id_copy = object::uid_to_inner(&project_id);
        
        let project = Project {
            id: project_id,
            creator,
            name: string::utf8(name),
            description: string::utf8(description),
            funding_goal,
            total_raised: balance::zero<SUI>(),
            donor_count: 0,
        };

        event::emit(ProjectCreated {
            project_id: project_id_copy,
            creator,
            name: project.name,
            funding_goal,
        });

        transfer::public_transfer(project, creator);
    }

    public entry fun donate(
        project: &mut Project,
        payment: Coin<SUI>,
        ctx: &mut TxContext
    ) {
        let amount = coin::value(&payment);
        assert!(amount > 0, EInsufficientFunds);

        let donor = tx_context::sender(ctx);
        let project_id = object::uid_to_inner(&project.id);
        
        // Add the payment to the project's balance
        let payment_balance = coin::into_balance(payment);
        balance::join(&mut project.total_raised, payment_balance);
        
        // Increment donor count
        project.donor_count = project.donor_count + 1;

        // Create donation receipt NFT
        let receipt_id = object::new(ctx);
        let receipt_id_copy = object::uid_to_inner(&receipt_id);
        
        let receipt = DonationReceipt {
            id: receipt_id,
            project_id,
            donor,
            amount,
        };

        event::emit(DonationMade {
            project_id,
            donor,
            amount,
            receipt_id: receipt_id_copy,
        });

        transfer::public_transfer(receipt, donor);
    }

    public entry fun withdraw(
        project: &mut Project,
        ctx: &mut TxContext
    ) {
        let creator = tx_context::sender(ctx);
        assert!(project.creator == creator, ENotProjectCreator);
        
        let balance_amount = balance::value(&project.total_raised);
        assert!(balance_amount > 0, EInsufficientBalance);
        
        let withdrawn_balance = balance::split(&mut project.total_raised, balance_amount);
        let withdrawn_coin = coin::from_balance(withdrawn_balance, ctx);
        
        let project_id = object::uid_to_inner(&project.id);
        
        event::emit(FundsWithdrawn {
            project_id,
            creator,
            amount: balance_amount,
        });

        transfer::public_transfer(withdrawn_coin, creator);
    }

    // Public view functions
    public fun get_project_info(project: &Project): (String, String, u64, u64, u64, address) {
        (
            project.name,
            project.description,
            project.funding_goal,
            balance::value(&project.total_raised),
            project.donor_count,
            project.creator
        )
    }

    public fun get_donation_receipt_info(receipt: &DonationReceipt): (ID, address, u64) {
        (receipt.project_id, receipt.donor, receipt.amount)
    }

    public fun get_project_id(project: &Project): ID {
        object::uid_to_inner(&project.id)
    }

    public fun get_receipt_id(receipt: &DonationReceipt): ID {
        object::uid_to_inner(&receipt.id)
    }
}
