class User < ApplicationRecord
  # Insecure: Allows mass assignment of all attributes, which can lead to privilege escalation or unauthorized data modification.
  # <expect-error> unsafe direct assignment
  attr_accessible :all
end


#Safe
class User < ApplicationRecord
  # Removed attr_accessible (deprecated and unsafe)
end


class UsersController < ApplicationController
  before_action :set_user, only: [:show, :update, :destroy]

  def create
    @user = User.new(user_params)
    if @user.save
      render json: @user, status: :created
    else
      render json: @user.errors, status: :unprocessable_entity
    end
  end

  def update
    if @user.update(user_params)
      render json: @user
    else
      render json: @user.errors, status: :unprocessable_entity
    end
  end

  private

  def set_user
    @user = User.find(params[:id])
  end

  # Strong parameters: Only permits safe attributes.
  def user_params
    params.require(:user).permit(:name, :email, :password)
  end
end
