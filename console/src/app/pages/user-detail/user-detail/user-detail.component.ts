import { Location } from '@angular/common';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatDialog } from '@angular/material/dialog';
import { ActivatedRoute, Params } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { ChangeType } from 'src/app/modules/changes/changes.component';
import {
    Gender,
    NotificationType,
    PasswordComplexityPolicy,
    UserAddress,
    UserEmail,
    UserPhone,
    UserProfile,
    UserState,
    UserView,
} from 'src/app/proto/generated/management_pb';
import { AuthUserService } from 'src/app/services/auth-user.service';
import { MgmtUserService } from 'src/app/services/mgmt-user.service';
import { OrgService } from 'src/app/services/org.service';
import { ToastService } from 'src/app/services/toast.service';

import { CodeDialogComponent } from '../code-dialog/code-dialog.component';

function passwordConfirmValidator(c: AbstractControl): any {
    if (!c.parent || !c) {
        return;
    }
    const pwd = c.parent.get('password');
    const cpwd = c.parent.get('confirmPassword');

    if (!pwd || !cpwd) {
        return;
    }
    if (pwd.value !== cpwd.value) {
        return { invalid: true, notequal: 'Password is not equal' };
    }
}

@Component({
    selector: 'app-user-detail',
    templateUrl: './user-detail.component.html',
    styleUrls: ['./user-detail.component.scss'],
})
export class UserDetailComponent implements OnInit, OnDestroy {
    public user!: UserView.AsObject;
    // public email: UserEmail.AsObject = { email: '' } as any;
    // public phone: UserPhone.AsObject = { phone: '' } as any;
    public address: UserAddress.AsObject = { id: '' } as any;
    public genders: Gender[] = [Gender.GENDER_MALE, Gender.GENDER_FEMALE, Gender.GENDER_DIVERSE];
    public languages: string[] = ['de', 'en'];

    public passwordForm!: FormGroup;
    public addressForm!: FormGroup;

    public isMgmt: boolean = false;
    private subscription: Subscription = new Subscription();
    public emailEditState: boolean = false;
    public phoneEditState: boolean = false;

    public ChangeType: any = ChangeType;
    public loading: boolean = false;

    public UserState: any = UserState;
    public policy!: PasswordComplexityPolicy.AsObject;
    constructor(
        public translate: TranslateService,
        private route: ActivatedRoute,
        private toast: ToastService,
        private mgmtUserService: MgmtUserService,
        private fb: FormBuilder,
        private _location: Location,
        private dialog: MatDialog,
        private orgService: OrgService,
        public authUserService: AuthUserService,
    ) {
        const validators: Validators[] = [Validators.required];

        this.orgService.GetPasswordComplexityPolicy().then(data => {
            this.policy = data.toObject();
            if (this.policy.minLength) {
                validators.push(Validators.minLength(this.policy.minLength));
            }
            if (this.policy.hasLowercase) {
                validators.push(Validators.pattern(/[a-z]/g));
            }
            if (this.policy.hasUppercase) {
                validators.push(Validators.pattern(/[A-Z]/g));
            }
            if (this.policy.hasNumber) {
                validators.push(Validators.pattern(/[0-9]/g));
            }
            if (this.policy.hasSymbol) {
                // All characters that are not a digit or an English letter \W or a whitespace \S
                validators.push(Validators.pattern(/[\W\S]/));
            }

            this.passwordForm = this.fb.group({
                password: ['', validators],
                confirmPassword: ['', [...validators, passwordConfirmValidator]],
            });
        }).catch(error => {
            console.log('no password complexity policy defined!');
            this.passwordForm = this.fb.group({
                password: ['', []],
                confirmPassword: ['', [passwordConfirmValidator]],
            });
        });

        this.addressForm = this.fb.group({
            streetAddress: [''],
            postalCode: [''],
            locality: [''],
            region: [''],
            country: [''],
        });
    }

    public ngOnInit(): void {
        this.subscription = this.route.params.subscribe(params => {
            this.loading = true;
            this.getData(params).then(() => {
                this.loading = false;
            }).catch(error => {
                this.loading = false;
            });
        });
    }

    public ngOnDestroy(): void {
        this.subscription.unsubscribe();
    }

    public deletePhone(): void {
        this.user.phone = '';
        this.savePhone();
    }

    public enterCode(): void {
        const dialogRef = this.dialog.open(CodeDialogComponent, {
            data: {
                number: this.user.phone,
            },
        });

        dialogRef.afterClosed().subscribe(code => {
            if (code) {
                this.toast.showInfo('TODO: implement service');
            }
        });
    }

    public saveProfile(profileData: UserProfile.AsObject): void {
        this.user.firstName = profileData.firstName;
        this.user.lastName = profileData.lastName;
        this.user.nickName = profileData.nickName;
        this.user.displayName = profileData.displayName;
        this.user.gender = profileData.gender;
        this.user.preferredLanguage = profileData.preferredLanguage;
        console.log(this.user);
        this.mgmtUserService
            .SaveUserProfile(this.user)
            .then((data: UserProfile) => {
                this.toast.showInfo('Saved Profile');
                this.user = Object.assign(this.user, data.toObject());
            })
            .catch(data => {
                this.toast.showError(data.message);
            });
    }

    public resendVerification(): void {
        console.log('resendverification');
        this.mgmtUserService.ResendEmailVerification(this.user.id).then(() => {
            this.toast.showInfo('Email was successfully sent!');
        }).catch(data => {
            this.toast.showError(data.message);
        });
    }

    public resendPhoneVerification(): void {
        this.mgmtUserService.ResendPhoneVerification(this.user.id).then(() => {
            this.toast.showInfo('Phoneverification was successfully sent!');
        }).catch(data => {
            this.toast.showError(data.message);
        });
    }

    public setInitialPassword(): void {
        if (this.passwordForm.valid && this.password && this.password.value) {
            this.mgmtUserService.SetInitialPassword(this.user.id, this.password.value).then((data: any) => {
                this.toast.showInfo('Set initial Password');
                this.user.email = data.toObject();
            }).catch(data => {
                this.toast.showError(data.message);
            });
        }
    }

    public sendSetPasswordNotification(): void {
        this.mgmtUserService.SendSetPasswordNotification(this.user.id, NotificationType.NOTIFICATIONTYPE_EMAIL)
            .then((data: any) => {
                this.toast.showInfo('Set initial Password');
                this.user.email = data.toObject();
            }).catch(data => {
                this.toast.showError(data.message);
            });
    }

    public saveEmail(): void {
        this.emailEditState = false;
        this.mgmtUserService
            .SaveUserEmail(this.user.id, this.user.email).then((data: UserEmail) => {
                this.toast.showInfo('Saved Email');
                this.user.email = data.toObject().email;
            }).catch(data => {
                this.toast.showError(data.message);
            });
    }

    public savePhone(): void {
        this.phoneEditState = false;
        this.mgmtUserService
            .SaveUserPhone(this.user.id, this.user.phone).then((data: UserPhone) => {
                this.toast.showInfo('Saved Phone');
                this.user.phone = data.toObject().phone;
            }).catch(data => {
                this.toast.showError(data.message);
            });
    }

    public saveAddress(): void {
        if (!this.address.id) {
            this.address.id = this.user.id;
        }

        this.address.streetAddress = this.streetAddress?.value;
        this.address.postalCode = this.postalCode?.value;
        this.address.locality = this.locality?.value;
        this.address.region = this.region?.value;
        this.address.country = this.country?.value;

        this.mgmtUserService
            .SaveUserAddress(this.address as UserAddress.AsObject).then((data: UserAddress) => {
                this.toast.showInfo('Saved Address');
                this.address = data.toObject();
            }).catch(data => {
                this.toast.showError(data.message);
            });
    }

    public navigateBack(): void {
        this._location.back();
    }

    public get streetAddress(): AbstractControl | null {
        return this.addressForm.get('streetAddress');
    }
    public get postalCode(): AbstractControl | null {
        return this.addressForm.get('postalCode');
    }
    public get locality(): AbstractControl | null {
        return this.addressForm.get('locality');
    }
    public get region(): AbstractControl | null {
        return this.addressForm.get('region');
    }
    public get country(): AbstractControl | null {
        return this.addressForm.get('country');
    }

    public get password(): AbstractControl | null {
        return this.passwordForm.get('password');
    }
    public get confirmPassword(): AbstractControl | null {
        return this.passwordForm.get('confirmPassword');
    }

    private async getData({ id }: Params): Promise<void> {
        this.isMgmt = true;
        this.mgmtUserService.GetUserByID(id).then(user => {
            console.log(user.toObject());
            this.user = user.toObject();
        }).catch(err => {
            console.error(err);
        });
        // this.profile = (await this.mgmtUserService.GetUserProfile(id)).toObject();
        // this.email = (await this.mgmtUserService.GetUserEmail(id)).toObject();
        // this.phone = (await this.mgmtUserService.GetUserPhone(id)).toObject();
        this.address = (await this.mgmtUserService.GetUserAddress(id)).toObject();
        this.addressForm.patchValue(this.address);
    }
}